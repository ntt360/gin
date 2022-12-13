package config

import (
	"errors"
	"github.com/ntt360/gin/core/gvalue"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

// app env type
const (
	Prod  = "prod"  // 线上
	Stage = "stage" // 预上线
	Test  = "test"  // 测试
	Dev   = "dev"   // 开发
)

// Model app global config model.
type Model struct {
	Env          string           `yaml:"env" env:"GINX_ENV"`
	LocalIP      string           `yaml:"-"`
	Hostname     string           `yaml:"-"`
	IdcName      string           `yaml:"-"`
	Name         string           `yaml:"name" env:"GINX_NAME"`
	Summary      string           `yaml:"summary" env:"GINX_SUMMARY"`
	HomeDir      string           `yaml:"home_dir" env:"GINX_HOME_DIR"`
	Mysql        mysqlDbModel     `yaml:"mysql_db"`
	Elastic      elasticSearch    `yaml:"elastic"`
	Redis        redisConfig      `yaml:"redis"`
	WebServerLog webServerLogConf `yaml:"web_server_log"`
	Log          logConf          `yaml:"log"`
	Grpc         grpcConfig       `yaml:"grpc"`
	HTTP         httpConfig       `yaml:"http"`
	HTTPS        httpsConfig      `yaml:"https"`
	ServerName   []string         `yaml:"server_name" env:"GINX_SERVER_NAME"`
	Task         taskConfig       `yaml:"task"`
	Tpl          tpl              `yaml:"tpl"`
	Auth         auth             `yaml:"auth"`
	Trace        Trace            `yaml:"trace"`
	PProf        pprof            `yaml:"pprof"`
	Metrics      metrics          `yaml:"metrics"`
	Cors         cors             `yaml:"cors"`

	APICallbackRegExp string `yaml:"api_callback_reg_exp"`
}

// MergeEnv merge yaml config and linux env same var.
func (m *Model) MergeEnv() {
	assign(reflect.ValueOf(m))
}

// Init 配置文件初始化
func Init(prjHome string) (Model, error) {
	configFile := strings.TrimRight(prjHome, "/") + "/config/config.yaml"
	c, err := os.Stat(configFile)
	var conf Model
	if !errors.Is(err, os.ErrNotExist) && c != nil {
		data, err := os.ReadFile(configFile)
		if err != nil {
			return conf, err
		}

		err = yaml.Unmarshal(data, &conf)
		if err != nil {
			return conf, err
		}
	}

	// 合并环境变量值，环境变量对应值优先级高于配置文件
	conf.HomeDir = strings.TrimRight(prjHome, "/")

	conf.MergeEnv()

	// 内部一些启动常量
	conf.LocalIP = gvalue.LocalIP()
	conf.IdcName = gvalue.IdcName()
	conf.Hostname = gvalue.Hostname()

	return conf, nil
}

func assign(v reflect.Value) {
	v = reflect.Indirect(v)
	for i := 0; i < v.NumField(); i++ {
		envKey := v.Type().Field(i).Tag.Get("env")
		fEnvVal, keyExit := os.LookupEnv(envKey)
		processOne(fEnvVal, keyExit, v.Field(i))
	}
}

func processOne(fEnvVal string, envKeyExist bool, vItem reflect.Value) {
	if !vItem.CanSet() {
		return
	}

	switch vItem.Type().Kind() {
	case reflect.String:
		if envKeyExist {
			vItem.SetString(fEnvVal)
		}
	case reflect.Int, reflect.Int64, reflect.Int32:
		eVal, e := strconv.ParseInt(fEnvVal, 0, vItem.Type().Bits())
		if e == nil && envKeyExist {
			vItem.SetInt(eVal)
		}
	case reflect.Bool:
		eVal, e := strconv.ParseBool(fEnvVal)
		if e == nil && envKeyExist {
			vItem.SetBool(eVal)
		}
	case reflect.Float32, reflect.Float64:
		eVal, e := strconv.ParseFloat(fEnvVal, vItem.Type().Bits())
		if e == nil && envKeyExist {
			vItem.SetFloat(eVal)
		}
	case reflect.Slice:
		eVals := strings.Split(fEnvVal, ",")
		if len(eVals) <= 0 || !envKeyExist {
			break
		}
		sl := reflect.MakeSlice(vItem.Type(), len(eVals), len(eVals))
		for key, val := range eVals {
			processOne(val, envKeyExist, sl.Index(key))
		}
		vItem.Set(sl)
	case reflect.Struct:
		assign(vItem)
	default:
	}
}

// GetHTTPSCertFile get project https cert file content.
func (m Model) GetHTTPSCertFile() string {
	if path.IsAbs(m.HTTPS.CertFile) {
		return m.HTTPS.CertFile
	}

	return m.HomeDir + "/" + m.HTTPS.CertFile
}

// GetHTTPSKeyFile get https cert key file content.
func (m Model) GetHTTPSKeyFile() string {
	if path.IsAbs(m.HTTPS.CertKey) {
		return m.HTTPS.CertKey
	}

	return m.HomeDir + "/" + m.HTTPS.CertKey
}
