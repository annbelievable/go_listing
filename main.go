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

	router.Handle("/admin-register", http.HandlerFunc(AdminRegister)).Methods("GET")
	router.Handle("/admin-register", parseFormHandler(http.HandlerFunc(AdminRegisterAction))).Methods("POST")
	router.Handle("/admin-login", http.HandlerFunc(AdminLogin)).Methods("GET")
	router.Handle("/admin-login", parseFormHandler(http.HandlerFunc(AdminLoginAction))).Methods("POST")
	router.Handle("/admin-homepage", http.HandlerFunc(AdminHomepage)).Methods("GET")
	router.HandleFunc("/admin-logout", AdminLogout).Methods("POST")

	router.HandleFunc("/test", Test).Methods("GET")
	router.HandleFunc("/bad-request", BadRequest).Methods("GET")
	router.HandleFunc("/access-denied", AccessDenied).Methods("GET")
	router.HandleFunc("/500", InternalServerError).Methods("GET")
	router.NotFoundHandler = notFound()

	router.Use(recoverHandler)
	router.Use(loggingHandler)
	router.Use(sessionHandler)

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
		LogError(err)
		InternalServerError(w, r)
		return
	}

	err = database.InsertAdmin(db, email, hashedPwd)

	if err != nil {
		LogError(err)
		InternalServerError(w, r)
		return
	}

	http.Redirect(w, r, "/admin-login", http.StatusFound)
}

func AdminLoginAction(w http.ResponseWriter, r *http.Request) {
	email := r.Form.Get("email")
	password := r.Form.Get("password")
	admin, err := database.SelectAdmin(db, email)

	if err != nil {
		LogError(err)
		InternalServerError(w, r)
		return
	}

	match := handlers.ComparePasswords(admin.Password, password)

	if !match {
		ctx := r.Context()
		ctx = context.WithValue(r.Context(), "Message", "Login failed.")
		AdminLogin(w, r.WithContext(ctx))
		return
	}

	sessionId := uuid.NewString()
	expiryDate := time.Now().Add(30 * time.Minute)

	database.DeleteAdminSessionByAdminId(db, admin.Id)
	err = database.InsertAdminSession(db, sessionId, admin.Id, expiryDate)
	if err != nil {
		LogError(err)
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

func AdminLogout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("session_id")
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	sessionId := c.Value
	session, err := database.SelectAdminSession(db, sessionId)

	if err != nil {
		LogError(err)
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if len(session.SessionId) == 0 && session.ExpiryDate.Before(time.Now()) {
		database.DeleteAdminSession(db, sessionId)
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_id",
		Value:   "",
		Expires: time.Now(),
	})

	http.Redirect(w, r, "/", http.StatusFound)
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
	render(w, data)
}

func AdminRegister(w http.ResponseWriter, r *http.Request) {
	ctxVal := r.Context().Value("LoggedIn")
	if ctxVal != nil {
		loggedIn := ctxVal.(bool)
		if loggedIn {
			http.Redirect(w, r, "/admin-homepage", http.StatusFound)
		}
	}

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
	ctxVal := r.Context().Value("LoggedIn")
	if ctxVal != nil {
		loggedIn := ctxVal.(bool)
		if loggedIn {
			http.Redirect(w, r, "/admin-homepage", http.StatusFound)
		}
	}

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
	ctxVal := r.Context().Value("LoggedIn")
	if ctxVal != nil {
		loggedIn := ctxVal.(bool)
		if !loggedIn {
			http.Redirect(w, r, "/admin-login", http.StatusFound)
		}
	}

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
		log.Printf("[INFO] NOT FOUND - %s %s %q\n", GetIP(r), r.Method, r.URL.String())
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
		log.Printf("[INFO] %s %s %q %v\n", GetIP(r), r.Method, r.URL.String(), t2.Sub(t1))
	})
}

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-FORWARDED-FOR")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

func recoverHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				LogError(err.(error))
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
			LogError(err)
			InternalServerError(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func sessionHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		c, err := r.Cookie("session_id")

		if err != nil {
			ctx = context.WithValue(r.Context(), "LoggedIn", false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		sessionId := c.Value
		session, err := database.SelectAdminSession(db, sessionId)

		if err != nil || len(session.SessionId) == 0 {
			if len(session.SessionId) > 0 {
				LogError(err)
			}
			ctx = context.WithValue(r.Context(), "LoggedIn", false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		database.DeleteAdminSessionByAdminId(db, session.AdminUser)
		newSessionId := uuid.NewString()
		expiryDate := time.Now().Add(30 * time.Minute)
		err = database.InsertAdminSession(db, newSessionId, session.AdminUser, expiryDate)

		if err != nil {
			LogError(err)
			ctx = context.WithValue(r.Context(), "LoggedIn", false)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:    "session_id",
			Value:   newSessionId,
			Expires: expiryDate,
		})

		ctx = context.WithValue(r.Context(), "LoggedIn", true)
		next.ServeHTTP(w, r.WithContext(ctx))
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
		LogError(err)
		http.Error(w, "Something went wrong.", 500)
	}
}

func LogError(err error) {
	log.Printf("[ERROR] %+v\n", err)
}
