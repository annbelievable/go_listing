package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"go_my_diary/database"
	"go_my_diary/handlers"

	"github.com/gorilla/mux"
)

var templates = template.Must(template.ParseGlob("./views/*.html"))
var pageTemplates = template.Must(template.ParseGlob("./views/pages/*"))
var db *sql.DB

type MessageData struct {
	Message string
	Errors  map[string]string
}

type PageContent struct {
	Title   string
	Content string
}

func main() {
	db = database.ConnectDatabase()
	err := database.CreateAdminTable(db)
	// server should still show some pages despite no data from database
	if err != nil {
		log.Println(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/", Homepage).Methods("GET")
	router.HandleFunc("/admin-register", AdminRegister).Methods("GET")
	router.HandleFunc("/admin-register", AdminRegisterAction).Methods("POST")
	router.HandleFunc("/admin-login", AdminLogin).Methods("GET")
	router.HandleFunc("/admin-login", AdminLoginAction).Methods("POST")
	router.HandleFunc("/admin-homepage", AdminHomepage).Methods("GET")
	router.HandleFunc("/admin-logout", AdminLogout).Methods("POST")
	router.HandleFunc("/test", Test).Methods("GET")
	router.HandleFunc("/bad-request", BadRequest).Methods("GET")
	router.HandleFunc("/access-denied", AccessDenied).Methods("GET")
	router.HandleFunc("/500", InternalServerError).Methods("GET")
	router.NotFoundHandler = notFound()

	router.Use(recoverHandler)
	router.Use(loggingHandler)

	log.Println("Starting server")
	log.Fatal(http.ListenAndServe(":8080", router))
}

// actions
func AdminRegisterAction(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")
	exist, err := database.AdminEmailExist(db, email)

	if err != nil {
		log.Println(err.Error)
		renderPage(w, "internal_server_error.html", nil)
		return
	}

	if exist > 0 {
		renderPage(w, "admin_register.html", MessageData{Message: "Email already exist."})
		return
	}

	hashedPwd, err := handlers.HashAndSalt(password)

	if err != nil {
		log.Println(err.Error)
		renderPage(w, "internal_server_error.html", nil)
		return
	}

	err = database.InsertAdmin(db, email, hashedPwd)

	if err != nil {
		log.Println(err.Error)
		renderPage(w, "internal_server_error.html", nil)
		return
	}

	renderPage(w, "admin_login.html", MessageData{Message: "Registration success."})
}

func AdminLoginAction(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		log.Println(err.Error)
		renderPage(w, "internal_server_error.html", nil)
		return
	}

	email := r.Form.Get("email")
	password := r.Form.Get("password")

	hpwd, err := database.SelectAdminHpwd(db, email)

	if err != nil {
		log.Println(err.Error)
		renderPage(w, "internal_server_error.html", nil)
		return
	}

	match := handlers.ComparePasswords(hpwd, password)

	if !match {
		renderPage(w, "admin_login.html", MessageData{Message: "Login failed."})
		return
	}

	renderPage(w, "admin_homepage.html", MessageData{Message: "Successfully login."})
}

// simple views
func Homepage(w http.ResponseWriter, r *http.Request) {
	data := PageContent{
		Title:   "Homepage",
		Content: "Homepage content for users to see",
	}

	render(w, data)
}

func AdminRegister(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "admin_register.html", nil)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "admin_login.html", nil)
}

func AdminHomepage(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "admin_homepage.html", nil)
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "homepage.html", nil)
}

// 500
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	renderPage(w, "internal_server_error.html", nil)
}

// 400
func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	renderPage(w, "bad_request.html", nil)
}

// 401
func AccessDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	renderPage(w, "access_denied.html", nil)
}

// 404
func notFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		renderPage(w, "notfound.html", nil)
	})
}

// MIDDLEWARE

func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t1 := time.Now()
		next.ServeHTTP(w, r)
		t2 := time.Now()
		log.Printf("[INFO] %s %q %v\n", r.Method, r.URL.String(), t2.Sub(t1))
	})
}

func recoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("[ERROR] %+v\n", err)
				http.Error(w, "Something went wrong.", 500)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// UTIL FUNC

// this is for testing purpose only
func Test(w http.ResponseWriter, r *http.Request) {
	renderPage(w, "test.html", nil)
}

// general page rendering
// potentially can pass in title, teaser, content, general message
func render(w http.ResponseWriter, data interface{}) {
	err := templates.ExecuteTemplate(w, "layout.html", data)
	checkError(w, err)
}

func renderPage(w http.ResponseWriter, fileName string, data interface{}) {
	err := pageTemplates.ExecuteTemplate(w, fileName, data)
	checkError(w, err)
}

func checkError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("[ERROR] %+v\n", err)
		http.Error(w, "Something went wrong.", 500)
	}
}
