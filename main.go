package main

import (
	"context"
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/annbelievable/go_listing/database"
	"github.com/annbelievable/go_listing/handlers"
	"github.com/annbelievable/go_listing/models"
	"github.com/google/uuid"

	"github.com/gorilla/mux"
)

var templates = template.Must(template.ParseGlob("./views/**/*.html"))
var db *sql.DB

func main() {
	db = database.ConnectDatabase()

	router := mux.NewRouter()
	router.HandleFunc("/", Homepage).Methods("GET")
	router.HandleFunc("/admin-register", AdminRegister).Methods("GET")
	router.Handle("/admin-register", parseFormHandler(http.HandlerFunc(AdminRegisterAction))).Methods("POST")
	router.HandleFunc("/admin-login", AdminLogin).Methods("GET")
	router.Handle("/admin-login", parseFormHandler(http.HandlerFunc(AdminLoginAction))).Methods("POST")
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
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	exist := database.AdminEmailExist(db, email)

	if exist {
		ctx := r.Context()
		ctx = context.WithValue(r.Context(), "Message", "Email already registered.")
		AdminRegister(w, r.WithContext(ctx))
		return
	}

	hashedPwd, err := handlers.HashAndSalt(password)

	if err != nil {
		log.Println(err.Error)
		InternalServerError(w, r)
		return
	}

	err = database.InsertAdmin(db, email, hashedPwd)

	if err != nil {
		log.Println(err.Error)
		InternalServerError(w, r)
		return
	}

	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func AdminLoginAction(w http.ResponseWriter, r *http.Request) {
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	hpwd, err := database.SelectAdminHpwd(db, email)

	if err != nil {
		log.Println(err.Error)
		InternalServerError(w, r)
		return
	}

	match := handlers.ComparePasswords(hpwd, password)

	if !match {
		ctx := r.Context()
		ctx = context.WithValue(r.Context(), "Message", "Login failed.")
		AdminLogin(w, r.WithContext(ctx))
		return
	}

	sessionId := uuid.NewString()
	expiryDate := time.Now().Add(30 * time.Minute)

	err = database.InsertAdminSession(db, sessionId, email, expiryDate)
	if err != nil {
		log.Println(err.Error)
		InternalServerError(w, r)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   sessionId,
		Expires: expiryDate,
	})

	http.Redirect(w, r, "/admin-homepage", http.StatusFound)
}

// simple views
func Homepage(w http.ResponseWriter, r *http.Request) {
	data := models.Page{
		Title:   "Home",
		Content: "This is My listing. Please enjoy browsing.",
	}
	ctxMsg := r.Context().Value("Message")
	if ctxMsg != nil {
		data.Message = ctxMsg.(string)
	}
	log.Printf("Message value %v\n", ctxMsg)
	render(w, data)
}

func AdminRegister(w http.ResponseWriter, r *http.Request) {
	data := models.Page{
		Title:   "Admin Registration",
		Content: "",
	}
	ctxMsg := r.Context().Value("Message")
	if ctxMsg != nil {
		data.Message = ctxMsg.(string)
	}
	renderPage(w, "admin_register.html", data)
}

func AdminLogin(w http.ResponseWriter, r *http.Request) {
	data := models.Page{
		Title:   "Admin Login",
		Content: "",
	}
	ctxMsg := r.Context().Value("Message")
	if ctxMsg != nil {
		data.Message = ctxMsg.(string)
	}
	renderPage(w, "admin_login.html", data)
}

func AdminHomepage(w http.ResponseWriter, r *http.Request) {
	data := models.Page{
		Title:   "Admin Homepage",
		Content: "",
	}
	ctxMsg := r.Context().Value("Message")
	if ctxMsg != nil {
		data.Message = ctxMsg.(string)
	}
	render(w, data)
}

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	data := models.Page{
		Title:   "Homepage",
		Content: "This is My listing. Please enjoy browsing.",
	}
	render(w, data)
}

// 500
func InternalServerError(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusInternalServerError)
	data := models.Page{
		Title:   "500: Internal Server Error",
		Content: "An error occurred, please contact admin about it.",
	}
	render(w, data)
}

// 400
func BadRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusBadRequest)
	data := models.Page{
		Title:   "400: Bad Request",
		Content: "Please try again.",
	}
	render(w, data)
}

// 401
func AccessDenied(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
	data := models.Page{
		Title:   "401: Access Denied",
		Content: "You're not allowed to access this content.",
	}
	render(w, data)
}

// 404
func notFound() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		data := models.Page{
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

func parseFormHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()

		if err != nil {
			log.Println(err.Error)
			InternalServerError(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isLoggedInHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie("session_id")
		if err != nil {
			if err == http.ErrNoCookie {
				AccessDenied(w, r)
				return
			}
			BadRequest(w, r)
			return
		}

		sessionId := c.Value
		session := database.SelectAdminSession(sessionId)

		if session != nil {
			AccessDenied(w, r)
			return
		}

		if session.ExpiryDate.After(time.Now()) {
			// delete session here
			AccessDenied(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// UTIL FUNC

// this is for testing purpose only
func Test(w http.ResponseWriter, r *http.Request) {
	data := models.Page{
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
