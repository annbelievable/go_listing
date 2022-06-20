package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/annbelievable/go_listing/database"
	"github.com/annbelievable/go_listing/handlers"

	"github.com/gorilla/mux"
)

var templates = template.Must(template.ParseGlob("./views/**/*.html"))
var db *sql.DB

type MessageData struct {
	Message string
	Errors  map[string]string
}

type PageData struct {
	Id      string
	Url     string
	Title   string
	Teaser  string
	Content string
	Message string
	Errors  map[string]string
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
	data := PageData{
		Title:   "Home",
		Content: "This is My listing. Please enjoy browsing.",
	}

	render(w, data)
}

func AdminRegister(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Admin Registration",
		Content: "",
	}
	renderPage(w, "admin_register.html", data)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Admin Login",
		Content: "",
	}
	renderPage(w, "admin_login.html", data)
}

func AdminHomepage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Admin Homepage",
		Content: "",
	}
	render(w, data)
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title:   "Homepage",
		Content: "This is My listing. Please enjoy browsing.",
	}
	render(w, data)
}

// 500
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	data := PageData{
		Title:   "500: Internal Server Error",
		Content: "An error occurred, please contact admin about it.",
	}
	render(w, data)
}

// 400
func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	data := PageData{
		Title:   "400: Bad Request",
		Content: "Please try again.",
	}
	render(w, data)
}

// 401
func AccessDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	data := PageData{
		Title:   "401: Access Denied",
		Content: "You're not allowed to access this content.",
	}
	render(w, data)
}

// 404
func notFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		data := PageData{
			Title:   "404: Not Found",
			Content: "The content youre looking for is not found.",
		}
		render(w, data)
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
	data := PageData{
		Title:   "Test",
		Content: "This is a test page.",
	}
	render(w, data)
}

// general page rendering
// potentially can pass in title, teaser, content, general message
func render(w http.ResponseWriter, data interface{}) {
	err := templates.ExecuteTemplate(w, "layout.html", data)
	checkError(w, err)
}

func renderPage(w http.ResponseWriter, fileName string, data interface{}) {
	err := templates.ExecuteTemplate(w, fileName, data)
	checkError(w, err)
}

func checkError(w http.ResponseWriter, err error) {
	if err != nil {
		log.Printf("[ERROR] %+v\n", err)
		http.Error(w, "Something went wrong.", 500)
	}
}
