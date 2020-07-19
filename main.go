package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
)

var store = sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
var sessionName = "session"

type Person struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Company string `json:"company"`
	City    string `json:"city"`
	Zip     string `json:"zip"`
	Geo     string `json:"geo"`
}

func loadData() (result []Person) {
	raw, _ := ioutil.ReadFile("./data.json")
	err := json.Unmarshal(raw, &result)
	if err != nil {
		log.Printf("err %v\n", err)
	}
	// log.Printf("data %v\n", result)
	return
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("home\n")
	target := os.Getenv("TITLE")
	if target == "" {
		target = "SimpleAuthWeb01"
	}
	t, err := template.ParseFiles("template/index.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	session, _ := store.Get(r, sessionName)
	username, ok := session.Values["username"].(string)
	if !ok {
		username = ""
	}
	people := loadData()
	err = t.Execute(w, struct {
		Title    string
		Username string
		People   []Person
	}{Title: target, Username: username, People: people})
	if err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Login\n")
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
			session, _ := store.Get(r, sessionName)
			session.Values["username"] = username
			_ = session.Save(r, w)
			http.Redirect(w, r, "/", 301)
		} else {
			message = "NG"
			w.WriteHeader(401)
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

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Logout\n")
	session, _ := store.Get(r, sessionName)
	session.Values["username"] = ""
	_ = session.Save(r, w)

	t, err := template.ParseFiles("template/logout.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	err = t.Execute(w, struct {
	}{})
	if err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

func personHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := strconv.Atoi(r.URL.Path[8:])
	log.Printf("person %v\n", i)
	session, _ := store.Get(r, sessionName)
	username := session.Values["username"]
	people := loadData()
	person := people[i]
	t, err := template.ParseFiles("template/base.html", "template/person.html")
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	params := map[string]interface{}{
		"Username": username,
		"Person":   person,
	}
	err = t.Execute(w, params)
	if err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

func main() {
	log.Print("SimpleAuthWeb01: starting server...")

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.Handle("/favicon.ico", fs)
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/person/", personHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("helloworld: listening on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
