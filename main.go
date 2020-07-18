package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"text/template"
)

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("helloworld: received a request")
	target := os.Getenv("TITLE")
	if target == "" {
		target = "SimpleAuthWeb01"
	}
	t, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	err = t.Execute(w, struct{ Title string }{Title: target})
	if err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

func main() {
	log.Print("helloworld: starting server...")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
