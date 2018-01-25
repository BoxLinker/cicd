package email

import (
	"net/http"
	"github.com/gorilla/context"
	"github.com/rs/cors"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/BoxLinker/boxlinker-api/pkg/amqp"
	"encoding/json"
	"time"
	"fmt"
	"net/smtp"
	"strings"
	"io/ioutil"
	"github.com/BoxLinker/boxlinker-api"
)


type EmailOption struct {
	User string
	UserTitle string
	Host string
	Password string
}

type ApiOptions struct {
	Listen string
	AMQPConsumer *amqp.Consumer
	EmailOption EmailOption
	TestMode bool
}

type Api struct {
	listen string
	amqpConsumer *amqp.Consumer
	emailOption EmailOption
	testMode bool
}

func NewApi(config ApiOptions) *Api {
	return &Api{
		listen: config.Listen,
		amqpConsumer: config.AMQPConsumer,
		emailOption: config.EmailOption,
		testMode: config.TestMode,
	}
}

func (a *Api) Run() error {
	cs := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"},
		AllowedHeaders: []string{"Origin", "Content-Type", "Accept", "token", "X-Requested-With", "X-Access-Token"},
	})

	//go sendEmail(a.amqpConsumer.NotifyMsg)

	globalMux := http.NewServeMux()

	emailRouter := mux.NewRouter()
	emailRouter.HandleFunc("/v1/email/sendTest", a.Send).Methods("POST")
	emailRouter.HandleFunc("/v1/email/send", a.Send).Methods("POST")
	globalMux.Handle("/v1/email/", emailRouter)

	s := &http.Server{
		Addr: a.listen,
		Handler: context.ClearHandler(cs.Handler(globalMux)),
	}
	log.Infof("Email Server listen on: %s", a.listen)
	return s.ListenAndServe()
}

func (a *Api) Send(w http.ResponseWriter, r *http.Request) {

	b, err := ioutil.ReadAll(r.Body)
	log.Debugf("<- %s", string(b))
	if err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	form := &SendForm{}
	if err := json.Unmarshal(b, form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_FORM_VALIDATE_ERR, nil, err.Error())
		return
	}

	a.DefaultForm(form)

	if err := sendMail(form); err != nil {
		boxlinker.Resp(w, boxlinker.STATUS_INTERNAL_SERVER_ERR, nil, err.Error())
		return
	}
	boxlinker.Resp(w, boxlinker.STATUS_OK, nil, "ok")
}


type SendForm struct {
	User string `json:"user"`
	UserTitle string `json:"userTitle"`
	Password string `json:"password"`
	Host string `json:"host"`
	To []string `json:"to"`

	Subject string `json:"subject"`
	Body string `json:"body"`
}

func (a *Api) DefaultForm(f *SendForm){
	op := a.emailOption
	if f.User == "" {
		f.User = op.User
	}
	if f.UserTitle == "" {
		f.UserTitle = op.UserTitle
	}
	if f.Host == "" {
		f.Host = op.Host
	}
	if f.Password == "" {
		f.Password = op.Password
	}
}

func sendMail(form *SendForm, mailType ...string) error{
	user := form.User
	if user == "" {
		return fmt.Errorf("User 不能为空")
	}
	userTitle := form.UserTitle
	if userTitle == "" {
		userTitle = user
	}
	password := form.Password
	if password == "" {
		return fmt.Errorf("Password 不能为空")
	}
	host := form.Host
	if host == "" {
		return fmt.Errorf("Host 不能为空")
	}
	if len(form.To) == 0 {
		return fmt.Errorf("To(邮件接受者) 不能为空")
	}
	to := strings.Join(form.To, ";")
	subject := form.Subject
	if subject == "" {
		return fmt.Errorf("Subject 不能为空")
	}
	body := form.Body
	mailTypeS := "html"

	if len(mailType) > 0 {
		mailTypeS = mailType[0]
	}

	hp := strings.Split(host, ":")
	auth := smtp.PlainAuth("", user, password, hp[0])
	var content_type string
	if mailTypeS == "html" {
		content_type = "Content-Type: text/"+ mailTypeS + "; charset=UTF-8"
	}else{
		content_type = "Content-Type: text/plain" + "; charset=UTF-8"
	}

	msg := []byte("To: " + to + "\r\nFrom: "+ userTitle +"<"+ user +">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + body)
	send_to := strings.Split(to, ";")
	err := smtp.SendMail(host, auth, user, send_to, msg)
	return err
}

func sendEmail(msgChan <-chan []byte){
	for b := range msgChan {
		log.Debugf("<- %s", string(b))
		form := &SendForm{}
		if err := json.Unmarshal(b, form); err != nil {
			log.Errorf("Json parse err: %s", err)
			return
		}
		if err := sendMail(form); err != nil {
			log.Errorf("Send email err: %s", err)
			return
		}
	}
}


func (a *Api) publishTest(form *SendForm) error {
	c := a.amqpConsumer
	i := 0
	for {
		time.Sleep(5*time.Second)
		//from := &mail.Address{c.String("mail-user-title"),c.String("mail-user")}
		//form := &SendForm{
		//	User: ,
		//	UserTitle: c.String("mail-user-title"),
		//	Host: c.String("mail-host"),
		//	Password: c.String("mail-password"),
		//	To: []string{"330785652@qq.com"},
		//	Subject: "Boxlinker 测试发送邮件",
		//	Body: "<h1>测试发送邮件内容，这个内容不长...</h1>",
		//}
		b, err := json.Marshal(form)
		if err != nil {
			return err
		}
		s := string(b)
		log.Debugf("-> %s", s)
		amqp.NewProducer(amqp.ProducerOptions{
			URI: c.URI,
			Exchange: c.Exchange,
			ExchangeType: c.ExchangeType,
			RoutingKey: c.QueueName,
		}).PublishOnce(s)
		i++
		if i >= 1 {
			break
		}
	}
	return nil
}
