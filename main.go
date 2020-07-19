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
	session, _ := store.Get(r, sessionName)
	people := loadData()
	renderPage(w, r, session, "index.html", people)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)
	switch r.Method {
	case http.MethodGet:
		renderPage(w, r, session, "login.html", map[string]interface{}{})
		break
	case http.MethodPost:
		r.ParseForm()
		username := r.Form["username"][0]
		password := r.Form["password"][0]
		hasher := md5.New()
		hasher.Write([]byte(username))
		md5password := hex.EncodeToString(hasher.Sum(nil))
		if password == md5password {
			session.Values["username"] = username
			_ = session.Save(r, w)
			http.Redirect(w, r, "/", 301)
			break
		}
		param := map[string]interface{}{
			"Message":  "Login Error",
			"Username": username,
		}
		w.WriteHeader(401)
		renderPage(w, r, session, "login.html", param)
		break
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, sessionName)
	session.Values["username"] = ""
	renderPage(w, r, session, "logout.html", nil)
}

func renderPage(w http.ResponseWriter, r *http.Request, session *sessions.Session, templateFilename string, param interface{}) {
	username := session.Values["username"]
	t, err := template.ParseFiles("template/base.html", "template/"+templateFilename)
	if err != nil {
		log.Fatalf("template error: %v", err)
	}
	params := map[string]interface{}{
		"Username": username,
		"Param":    param,
	}
	w.Header().Set("Content-type", "text/html")
	err = t.Execute(w, params)
	if err != nil {
		log.Printf("failed to execute template: %v", err)
	}
}

func personHandler(w http.ResponseWriter, r *http.Request) {
	i, _ := strconv.Atoi(r.URL.Path[8:])
	people := loadData()
	person := people[i]
	session, _ := store.Get(r, sessionName)
	renderPage(w, r, session, "person.html", person)
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
