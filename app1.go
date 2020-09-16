package main

import (
	"net/http"
	"log"
)

func main() {
	
	http.HandleFunc("/app1", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("app1"))
	})

	log.Fatal(http.ListenAndServe(":8081", nil))
}
