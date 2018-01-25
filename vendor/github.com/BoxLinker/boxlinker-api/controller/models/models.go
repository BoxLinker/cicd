package models

import (
	"github.com/go-xorm/xorm"
	"fmt"
	"net/url"
	"strings"
	log "github.com/Sirupsen/logrus"
	"github.com/go-xorm/core"
	_ "github.com/go-sql-driver/mysql"
	"github.com/urfave/cli"
)

//var (
//	tables []interface{}
//)
//
//func init(){
//	tables = append(tables, new(User), new(UserToBeConfirmed))
//}


type DBOptions struct {
	User string
	Password string
	Host string
	Port int
	Name string
}

func GetDBOptions(c *cli.Context) DBOptions {
	return DBOptions{
		User: 		c.String("db-user"),
		Password: 	c.String("db-password"),
		Host: 		c.String("db-host"),
		Port: 		c.Int("db-port"),
		Name: 		c.String("db-name"),
	}
}

func NewEngine(config DBOptions, t []interface{}) (*xorm.Engine, error){
	var Param string = "?"
	//var _tables []interface{}
	if strings.Contains(config.Name, Param) {
		Param = "&"
	}
	var connStr = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&loc=%s",
		url.QueryEscape(config.User),
		url.QueryEscape(config.Password),
		config.Host,
		config.Port,
		config.Name,"Asia%2FShanghai")

	log.Infof("Connect to db: %s", connStr)
	x, err := xorm.NewEngine("mysql", connStr)
	if err != nil {
		return nil,err
	}
	log.Info("Connect to db ok.")
	x.SetMapper(core.GonicMapper{})
	log.Infof("start to sync tables ...")
	//if len(t) > 0 {
	//	_tables = t[0]
	//} else {
	//	_tables = tables
	//}
	if err = x.StoreEngine("InnoDB").Sync2(t...); err != nil {
		return nil, fmt.Errorf("sync tables err: %v",err)
	}
	x.ShowSQL(true)
	return x, nil
}

