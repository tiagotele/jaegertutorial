# Jaeger tutorial

## Running Jaeger from cli

```
docker run --rm -p 6831:6831/udp -p 6832:6832/udp -p 16686:16686 jaegertracing/all-in-one:1.7 --log-level=debug
```

Access from browser here:

http://localhost:16686

## There are 3 apps that must be runned on separated tabs.
```
go run app1.go
go run app2.go
go run app3.go
```

Then access endpoint from app1
http://localhost:8081/app
or 
http://localhost:8081/app1to3


See the tracker on Jaeger.