package main

import (
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	otlog "github.com/opentracing/opentracing-go/log"
	xhttp "github.com/yurishkuro/opentracing-tutorial/go/lib/http"
	"github.com/opentracing/opentracing-go/ext"
	"io"
	"log"
	"time"
	"fmt"
	"net/http"
	"net/url"
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
	tracer, closer := initJaeger("app-distributed-2")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	helloTo := "My awesome string"

	http.HandleFunc("/app2", func(w http.ResponseWriter, r *http.Request) {
		
		time.Sleep(1 * time.Second)
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		span := tracer.StartSpan("format", ext.RPCServerOption(spanCtx))
		defer span.Finish()

		helloTo := r.FormValue("helloTo")
		helloStr := fmt.Sprintf("Hello, %s!", helloTo)
		span.LogFields(
			otlog.String("event", "string-format"),
			otlog.String("value", helloStr),
		)
		w.Write([]byte("app2"))
	})

	http.HandleFunc("/app2to3", func(w http.ResponseWriter, r *http.Request) {
		
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(r.Header))
		spanApp2to3 := tracer.StartSpan("format", ext.RPCServerOption(spanCtx))

		defer spanApp2to3.Finish()

		v := url.Values{}
		v.Set("helloTo", helloTo)
		url := "http://localhost:8083/app3"
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err.Error())
		}

		time.Sleep(1 * time.Second)
		ext.SpanKindRPCClient.Set(spanApp2to3)
		ext.HTTPUrl.Set(spanApp2to3, url)
		ext.HTTPMethod.Set(spanApp2to3, "GET")
		spanApp2to3.Tracer().Inject(
			spanApp2to3.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header),
		)

		resp, err := xhttp.Do(req)
		if err != nil {
			ext.LogError(spanApp2to3, err)
			panic(err.Error())
		}

		helloStr := string(resp)

		spanApp2to3.LogFields(
			otlog.String("event", "string-format"),
			otlog.String("value", helloStr),
		)

		w.Write([]byte("app2to3"))
	})

	log.Fatal(http.ListenAndServe(":8082", nil))
}
