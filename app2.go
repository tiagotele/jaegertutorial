package main

import (
	"net/http"
	"log"
)

func main() {

	http.HandleFunc("/app2", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("app2"))
	})

	log.Fatal(http.ListenAndServe(":8082", nil))
}
