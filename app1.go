package main

import (
	"context"
	"fmt"
	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
	xhttp "github.com/yurishkuro/opentracing-tutorial/go/lib/http"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/opentracing/opentracing-go/ext"
	"io"
	"log"
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
	tracer, closer := initJaeger("hello-world")
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	helloTo := "My awesome string"

	span := tracer.StartSpan("say-hello")
	span.SetTag("hello-to", helloTo)
	defer span.Finish()

	ctx := opentracing.ContextWithSpan(context.Background(), span)

	http.HandleFunc("/app1", func(w http.ResponseWriter, r *http.Request) {
		spanApp2, _ := opentracing.StartSpanFromContext(ctx, "printHello")

		defer spanApp2.Finish()

		v := url.Values{}
		v.Set("helloTo", helloTo)
		url := "http://localhost:8082/app2?" + v.Encode()
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err.Error())
		}

		ext.SpanKindRPCClient.Set(spanApp2)
		ext.HTTPUrl.Set(spanApp2, url)
		ext.HTTPMethod.Set(spanApp2, "GET")
		spanApp2.Tracer().Inject(
			spanApp2.Context(),
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header),
		)

		resp, err := xhttp.Do(req)
		if err != nil {
			ext.LogError(spanApp2, err)
			panic(err.Error())
		}

		helloStr := string(resp)

		spanApp2.LogFields(
			otlog.String("event", "string-format"),
			otlog.String("value", helloStr),
		)

		w.Write([]byte("app1"))
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
