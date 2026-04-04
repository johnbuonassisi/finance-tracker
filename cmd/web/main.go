package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {

	http.HandleFunc("/register", register)
	http.HandleFunc("/login", login)
	http.HandleFunc("/protected", protected)
	http.HandleFunc("/logout", logout)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func register(w http.ResponseWriter, r *http.Request) {
	userName := r.FormValue("username")
	if len(userName) < 8 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	password := r.FormValue("password")
	if len(password) < 8 {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, ok := users[userName]
	if ok {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	log.Printf("getting past here")

	hashedPassword, err := hashPassword(password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	users[userName] = &User{
		Name:           userName,
		HashedPassword: string(hashedPassword),
	}
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(hashedPassword), err
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func login(w http.ResponseWriter, r *http.Request) {
	userName := r.FormValue("username")
	password := r.FormValue("password")

	// does user exist
	user, ok := users[userName]
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	// does the incoming password match the hashed, stored password
	isMatch := checkPassword(user.HashedPassword, password)
	if !isMatch {
		http.Error(w, "incorrect password", http.StatusUnauthorized)
		return
	}

	// generate session and csrf tokens
	sessionToken, err := generateToken(32)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	csrfToken, err := generateToken(32)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// session cookie will be:
	// - sent by the browser with future requests
	// - removed from the browser in 24hrs
	// - sent in http requests and not accessible to frontend javascript code
	http.SetCookie(w, &http.Cookie{Name: "session",
		Value:    sessionToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
	})

	// cross site request forgery
	// - csrf token is available by client
	http.SetCookie(w, &http.Cookie{Name: "csrf",
		Value:    csrfToken,
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: false,
	})

	// store tokens
	user.SessionToken = sessionToken
	user.CSRFToken = csrfToken

	fmt.Fprintln(w, "Login success!")
}

func protected(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := authorize(r); err != nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	userName := r.FormValue("username")
	fmt.Fprintf(w, "Authorization check passed! Welcome %s!\n", userName)
}

func logout(w http.ResponseWriter, r *http.Request) {
	if err := authorize(r); err != nil {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: true,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     "csrf",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour),
		HttpOnly: false,
	})

	userName := r.FormValue("username")
	user, ok := users[userName]
	if !ok {
		http.Error(w, "not authorized", http.StatusNotFound)
	}
	user.SessionToken = ""
	user.CSRFToken = ""

	fmt.Fprintln(w, "Logged out successfully!")
}

func authorize(r *http.Request) error {
	userName := r.FormValue("username")
	user, ok := users[userName]
	if !ok {
		return errors.New("user not found")
	}

	sessionToken, err := r.Cookie("session")
	if err != nil {
		return errors.New("session cookie not found")
	}
	if sessionToken == nil || sessionToken.Value == "" || sessionToken.Value != user.SessionToken {
		return errors.New("session value token invalid")
	}

	csrfToken := r.Header.Get("csrf")
	if csrfToken == "" || csrfToken != user.CSRFToken {
		return errors.New("csrf token invalid")
	}

	return nil
}

func generateToken(length int) (string, error) {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type User struct {
	Name           string
	HashedPassword string
	SessionToken   string
	CSRFToken      string
}

var users = map[string]*User{}
