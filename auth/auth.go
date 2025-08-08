package auth

import (
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

var store = sessions.NewCookieStore([]byte("your-secret-key-change-this-in-production"))

func init() {
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
	}
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func GetSession(r *http.Request) (*sessions.Session, error) {
	return store.Get(r, "qr-linker-session")
}

func SaveSession(w http.ResponseWriter, r *http.Request, session *sessions.Session) error {
	return session.Save(r, w)
}

func SetUserSession(w http.ResponseWriter, r *http.Request, userID int, username string) error {
	session, err := GetSession(r)
	if err != nil {
		return err
	}

	session.Values["user_id"] = userID
	session.Values["username"] = username
	session.Values["authenticated"] = true

	return SaveSession(w, r, session)
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := GetSession(r)
	if err != nil {
		return err
	}

	session.Values["user_id"] = nil
	session.Values["username"] = nil
	session.Values["authenticated"] = false
	session.Options.MaxAge = -1

	return SaveSession(w, r, session)
}

func IsAuthenticated(r *http.Request) bool {
	session, err := GetSession(r)
	if err != nil {
		return false
	}

	auth, ok := session.Values["authenticated"].(bool)
	return ok && auth
}

func GetUserFromSession(r *http.Request) (int, string, bool) {
	session, err := GetSession(r)
	if err != nil {
		return 0, "", false
	}

	userID, ok1 := session.Values["user_id"].(int)
	username, ok2 := session.Values["username"].(string)
	
	if !ok1 || !ok2 {
		return 0, "", false
	}

	return userID, username, true
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !IsAuthenticated(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

