package main

import (
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("HI")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Print("handled message")
		fmt.Fprint(w, "Hi")
	})

	http.ListenAndServe(":8080", nil)
}
