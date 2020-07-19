package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var session_name = "session"

func homeHandler(w http.ResponseWriter, r *http.Request) {
	target := os.Getenv("TITLE")
	if target == "" {
		target = "SimpleAuthWeb01"
	}
	t, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	session, _ := store.Get(r, session_name)
	username, ok := session.Values["username"].(string)
	if !ok {
		username = ""
	}
	err = t.Execute(w, struct {
		Title    string
		Username string
	}{Title: target, Username: username})
	if err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/html")
	t, _ := template.ParseFiles("template/login.html")
	r.ParseForm()
	switch r.Method {
	case http.MethodPost:
		var message string
		var username string = r.Form["username"][0]
		var password string = r.Form["password"][0]
		hasher := md5.New()
		hasher.Write([]byte(username))
		tpassword := hex.EncodeToString(hasher.Sum(nil))
		log.Printf("tpass:%v\n", tpassword)
		if password == tpassword {
			session, _ := store.Get(r, session_name)
			session.Values["username"] = username
			_ = session.Save(r, w)
			http.Redirect(w, r, "/", 301)
		} else {
			message = "NG"
		}
		t.Execute(w, struct {
			Message  string
			Username string
		}{Message: message, Username: username})
		break
	case http.MethodGet:
		t.Execute(w, nil)
		break
	}
}

func main() {
	log.Print("helloworld: starting server...")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
