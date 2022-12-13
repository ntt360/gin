package jaeger

import (
	"github.com/ntt360/gin/core/config"
	"gorm.io/gorm"
)

const (
	callBackBeforeName = "jaeger:before"
	callBackAfterName  = "jaeger:after"
)

type Plugin struct {
	conf *config.Jaeger
}

func New(appConf *config.Jaeger) *Plugin {
	return &Plugin{
		conf: appConf,
	}
}

func (j *Plugin) Name() string {
	return "plugin:jaeger"
}

func (j *Plugin) IsAll() bool {
	return j.conf.LogSQL == "all"
}

func (j *Plugin) IsSlow() bool {
	return j.conf.LogSQL == "slow"
}

func (j *Plugin) Initialize(db *gorm.DB) error {
	err := db.Callback().Create().Before("*").Register(callBackBeforeName, before)
	if err != nil {
		return err
	}
	err = db.Callback().Query().Before("*").Register(callBackBeforeName, before)
	if err != nil {
		return err
	}
	err = db.Callback().Delete().Before("*").Register(callBackBeforeName, before)
	if err != nil {
		return err
	}
	err = db.Callback().Update().Before("*").Register(callBackBeforeName, before)
	if err != nil {
		return err
	}
	err = db.Callback().Row().Before("*").Register(callBackBeforeName, before)
	if err != nil {
		return err
	}
	err = db.Callback().Raw().Before("*").Register(callBackBeforeName, before)
	if err != nil {
		return err
	}

	err = db.Callback().Create().After("*").Register(callBackAfterName, after)
	if err != nil {
		return err
	}
	err = db.Callback().Query().After("*").Register(callBackAfterName, after)
	if err != nil {
		return err
	}
	err = db.Callback().Delete().After("*").Register(callBackAfterName, after)
	if err != nil {
		return err
	}
	err = db.Callback().Update().After("*").Register(callBackAfterName, after)
	if err != nil {
		return err
	}
	err = db.Callback().Row().After("*").Register(callBackAfterName, after)
	if err != nil {
		return err
	}
	err = db.Callback().Raw().After("*").Register(callBackAfterName, after)
	if err != nil {
		return err
	}

	return err
}
