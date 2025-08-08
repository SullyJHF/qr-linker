package main

import (
	"database/sql"
	"embed"
	"html/template"
	"log"
	"net/http"
	"qr-linker/auth"
	"qr-linker/database"
	"qr-linker/utils"
	"strings"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/*.css
var staticFS embed.FS

type PageData struct {
	Title     string
	Message   string
	URLs      []database.URL
	ShortURL  string
	Host      string
	Error     string
	Username  string
}

type LoginData struct {
	Title   string
	Error   string
	Message string
}

var db *database.DB

func main() {
	var err error
	db, err = database.NewDB("urls.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Public routes
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.Handle("/static/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/", publicRouteHandler)

	// Protected routes
	http.HandleFunc("/shorten", auth.RequireAuth(shortenHandler))

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func publicRouteHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	if path == "/" {
		// Homepage requires authentication
		if !auth.IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		homeHandler(w, r)
		return
	}
	
	// Short URL redirects are public
	shortHash := strings.TrimPrefix(path, "/")
	if shortHash != "" {
		redirectHandler(w, r, shortHash)
		return
	}
	
	http.NotFound(w, r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	// Set cache-control headers to prevent caching of dynamic content
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	
	tmpl, err := template.ParseFS(templatesFS, "templates/index.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		log.Printf("Template error: %v", err)
		return
	}

	urls, err := db.GetAllURLs()
	if err != nil {
		log.Printf("Error fetching URLs: %v", err)
		urls = []database.URL{}
	}

	// Get username from session
	_, username, _ := auth.GetUserFromSession(r)

	data := PageData{
		Title:    "QR Linker - URL Shortener",
		URLs:     urls,
		Host:     "http://localhost:8080",
		Username: username,
	}

	// Check for success parameter
	if success := r.URL.Query().Get("success"); success != "" {
		data.ShortURL = "/" + success
	}

	// Check for error parameter
	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		data.Error = errorMsg
	}

	if err := tmpl.Execute(w, data); err != nil {
		log.Printf("Render error: %v", err)
	}
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Check if already authenticated
		if auth.IsAuthenticated(r) {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		tmpl, err := template.ParseFS(templatesFS, "templates/login.html")
		if err != nil {
			http.Error(w, "Error loading template", http.StatusInternalServerError)
			return
		}

		data := LoginData{
			Title: "Login - QR Linker",
		}

		tmpl.Execute(w, data)
		return
	}

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			renderLoginError(w, "Invalid form data")
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		if username == "" || password == "" {
			renderLoginError(w, "Username and password are required")
			return
		}

		// Get user from database
		user, err := db.GetUserByUsername(username)
		if err != nil {
			if err == sql.ErrNoRows {
				renderLoginError(w, "Invalid username or password")
			} else {
				log.Printf("Database error: %v", err)
				renderLoginError(w, "An error occurred. Please try again.")
			}
			return
		}

		// Check password
		if !auth.CheckPasswordHash(password, user.PasswordHash) {
			renderLoginError(w, "Invalid username or password")
			return
		}

		// Set session
		err = auth.SetUserSession(w, r, user.ID, user.Username)
		if err != nil {
			log.Printf("Session error: %v", err)
			renderLoginError(w, "Failed to create session")
			return
		}

		// Redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	err := auth.ClearSession(w, r)
	if err != nil {
		log.Printf("Error clearing session: %v", err)
	}

	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func renderLoginError(w http.ResponseWriter, errorMsg string) {
	tmpl, err := template.ParseFS(templatesFS, "templates/login.html")
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}

	data := LoginData{
		Title: "Login - QR Linker",
		Error: errorMsg,
	}

	tmpl.Execute(w, data)
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Redirect(w, r, "/?error=Invalid+form+data", http.StatusSeeOther)
		return
	}

	fullURL := r.FormValue("url")
	if fullURL == "" {
		http.Redirect(w, r, "/?error=URL+is+required", http.StatusSeeOther)
		return
	}

	if !strings.HasPrefix(fullURL, "http://") && !strings.HasPrefix(fullURL, "https://") {
		fullURL = "https://" + fullURL
	}

	shortHash, err := utils.GenerateUniqueHash(db.CheckHashExists)
	if err != nil {
		log.Printf("Error generating hash: %v", err)
		http.Redirect(w, r, "/?error=Failed+to+generate+short+URL", http.StatusSeeOther)
		return
	}

	_, err = db.CreateURL(fullURL, shortHash)
	if err != nil {
		log.Printf("Error saving URL: %v", err)
		http.Redirect(w, r, "/?error=Failed+to+save+URL", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, "/?success="+shortHash, http.StatusSeeOther)
}

func redirectHandler(w http.ResponseWriter, r *http.Request, shortHash string) {
	// Set cache-control headers to prevent any caching of the redirect
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	
	url, err := db.GetURLByHash(shortHash)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = db.IncrementClicks(shortHash)
	if err != nil {
		log.Printf("Error incrementing clicks: %v", err)
	}

	http.Redirect(w, r, url.FullURL, http.StatusFound)
}

