package v1

type Config struct {
	Server struct {
		Addr string `yaml:"addr,omitempty"`
		Debug bool `yaml:"debug"`
	}    `yaml:"server,omitempty"`
	DB struct{
		Host string `yaml:"host,omitempty"`
		Port int `yaml:"port,omitempty"`
		User string `yaml:"user,omitempty"`
		Password string `yaml:"password,omitempty"`
		Name string `yaml:"name,omitempty"`
	} `yaml:"db,omitempty"`
}
