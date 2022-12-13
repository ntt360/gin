package config

import (
	"fmt"
	"golang.org/x/crypto/ssh"
)

// write db
type mysqlDbWModelItem struct {
	Host         string `yaml:"host" env:"GINX_MYSQL_WRITE_HOST"`
	Port         string `yaml:"port" env:"GINX_MYSQL_WRITE_PORT"`
	Database     string `yaml:"database" env:"GINX_MYSQL_WRITE_DATABASE"`
	Password     string `yaml:"password" env:"GINX_MYSQL_WRITE_PASSWORD"`
	Username     string `yaml:"username" env:"GINX_MYSQL_WRITE_USERNAME"`
	Timeout      int    `yaml:"timeout" env:"GINX_MYSQL_WRITE_TIMEOUT"`
	ReadTimeout  int    `yaml:"read_time_out" env:"GINX_MYSQL_WRITE_READ_TIMEOUT"`
	WriteTimeout int    `yaml:"write_time_out" env:"GINX_MYSQL_WRITE_WRITE_TIMEOUT"`
	Charset      string `yaml:"charset" env:"GINX_MYSQL_WRITE_CHARSET"`
	// connect pool
	ConnMaxLifeTime int `yaml:"conn_max_life_time" env:"GINX_MYSQL_WRITE_CONN_MAX_LIFE_TIME"`
	MaxIdleConns    int `yaml:"max_idle_conns" env:"GINX_MYSQL_WRITE_MAX_IDLE_CONNS"`
}

type sshConf struct {
	Enable   bool   `yaml:"enable"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

func (s *sshConf) DialWithPassword() (*ssh.Client, error) {
	address := fmt.Sprintf("%s:%d", s.Host, s.Port)
	config := &ssh.ClientConfig{
		User: s.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(s.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	return ssh.Dial("tcp", address, config)
}

// read db
type mysqlDbRModelItem struct {
	Host         string `yaml:"host" env:"GINX_MYSQL_READ_HOST"`
	Port         string `yaml:"port" env:"GINX_MYSQL_READ_PORT"`
	Database     string `yaml:"database" env:"GINX_MYSQL_READ_DATABASE"`
	Password     string `yaml:"password" env:"GINX_MYSQL_READ_PASSWORD"`
	Username     string `yaml:"username" env:"GINX_MYSQL_READ_USERNAME"`
	Timeout      int    `yaml:"timeout" env:"GINX_MYSQL_READ_TIMEOUT"`
	ReadTimeout  int    `yaml:"read_time_out" env:"GINX_MYSQL_READ_READ_TIMEOUT"`
	WriteTimeout int    `yaml:"write_time_out" env:"GINX_MYSQL_READ_WRITE_TIMEOUT"`
	Charset      string `yaml:"charset" env:"GINX_MYSQL_READ_CHARSET"`
	// connect pool
	ConnMaxLifeTime int `yaml:"conn_max_life_time" env:"GINX_MYSQL_READ_CONN_MAX_LIFE_TIME"`
	MaxIdleConns    int `yaml:"max_idle_conns" env:"GINX_MYSQL_READ_MAX_IDLE_CONNS"`
}

type Jaeger struct {
	LogSQL        string `yaml:"log_sql"`
	SlowQueryTime int    `yaml:"slow_query_time"`
}

type mysqlDbModel struct {
	Enable   bool              `yaml:"enable" env:"GINX_MTSQL_ENABLE"`
	Timezone string            `yaml:"timezone" env:"GINX_MYSQL_TIMEZONE"`
	LogMode  bool              `yaml:"log_mode" env:"GINX_MYSQL_LOG_MODE"`
	Write    mysqlDbWModelItem `yaml:"write"`
	Read     mysqlDbRModelItem `yaml:"read"`
	Jaeger   Jaeger            `yaml:"jaeger"`
	SSH      sshConf           `yaml:"ssh"`
}
