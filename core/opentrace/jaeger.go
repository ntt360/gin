package opentrace

import (
	"log"

	"github.com/ntt360/gin/core/config"
	jaegercfg "github.com/uber/jaeger-client-go/config"
	jaegerlog "github.com/uber/jaeger-client-go/log"
	"github.com/uber/jaeger-lib/metrics"
)

func Init(appConf *config.Model) error {
	conf, err := jaegercfg.FromEnv()
	if err != nil {
		return err
	}

	if len(appConf.Name) <= 0 {
		panic("the name do not allow empty, you must set it in config.yaml")
	}

	traceConf := appConf.Trace

	// 采样策略
	conf.Sampler.Type = traceConf.SampleType
	conf.Sampler.Param = traceConf.SampleParam
	conf.Reporter.LocalAgentHostPort = traceConf.AgentServer

	if appConf.Log.IsDebug() {
		log.Printf("tracer{ agent_server: %s, type: %s, param: %f } \n", traceConf.AgentServer, traceConf.SampleType, traceConf.SampleParam)
	}

	// Initialize tracer with a logger and a metrics factory
	_, err = conf.InitGlobalTracer(
		appConf.Name,
		jaegercfg.Logger(jaegerlog.NullLogger),
		jaegercfg.Metrics(metrics.NullFactory),
		jaegercfg.MaxTagValueLength(2000), // tag 最大运行存储2000个字符
	)

	if err != nil {
		return err
	}

	return nil
}
