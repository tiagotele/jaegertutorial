package main

import (
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/opentracing/opentracing-go/ext"
	"io"
	"log"
	"time"
	"fmt"
	"net/http"
)

func initJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans: true,
		},
	}
	tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
	if err != nil {
		panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
	}
	return tracer, closer
}

func main() {


	tracer, closer := initJaeger("app-distributed-3")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	http.HandleFunc("/app3", func(w http.ResponseWriter, r *http.Request) {
		
		
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		span := tracer.StartSpan("format", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		helloTo := r.FormValue("helloTo")
		helloStr := fmt.Sprintf("Hello, %s!", helloTo)
		span.LogFields(
			otlog.String("event", "string-format"),
			otlog.String("value", helloStr),
		)

		time.Sleep(5 * time.Second)

		w.Write([]byte("app3"))
	})

	log.Fatal(http.ListenAndServe(":8083", nil))
}
