package main

import (
	"embed"
	"html/template"
	"log"
	"net/http"
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
}

var db *database.DB

func main() {
	var err error
	db, err = database.NewDB("urls.db")
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	http.HandleFunc("/", routeHandler)
	http.HandleFunc("/shorten", shortenHandler)
	http.Handle("/static/", http.FileServer(http.FS(staticFS)))

	log.Println("Server starting on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func routeHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	
	if path == "/" {
		homeHandler(w, r)
		return
	}
	
	shortHash := strings.TrimPrefix(path, "/")
	if shortHash != "" {
		redirectHandler(w, r, shortHash)
		return
	}
	
	http.NotFound(w, r)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
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

	data := PageData{
		Title: "QR Linker - URL Shortener",
		URLs:  urls,
		Host:  "http://localhost:8080",
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
	url, err := db.GetURLByHash(shortHash)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	err = db.IncrementClicks(shortHash)
	if err != nil {
		log.Printf("Error incrementing clicks: %v", err)
	}

	http.Redirect(w, r, url.FullURL, http.StatusMovedPermanently)
}

