// package authplaceholder provides a quick and dirty username/password prompt for your prototype webapps.
package authplaceholder

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ContextKey is really just a string, to keep the context key for the unsername value unique.
type ContextKey string

// New creates a brand new PlaceholderAuth object from the configuration in the given filename.
//
// The file must exist, and must be valid JSON data, or an error is returned.
//
// The given mux will have POST requests to the validation path routed, as well as GET reqeusts to the void path.
func New(mux *http.ServeMux, filename string) (PlaceholderAuth, error) {
	pa, err := loadConfiguration(filename)
	if err != nil {
		return pa, fmt.Errorf("error loading placeholder auth configuration: %w", err)
	}
	mux.HandleFunc("POST "+pa.ValidatePath, pa.validateHandler)
	mux.HandleFunc("GET "+pa.VoidPath, pa.voidHandler)
	return pa, nil
}

func (pa PlaceholderAuth) verbose(format string, values ...any) {
	if !pa.Verbose {
		return
	}
	fmt.Printf("[PA] "+format, values...)
}

func (pa PlaceholderAuth) Wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pa.verbose("Wrapped handler happening. %s %s\n", r.Method, r.URL)
		cookie, err := r.Cookie(pa.CookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				pa.verbose("Coookie %q was not found. Sending auth form\n", pa.CookieName)
				pa.SendAuthForm(w)
				return
			} else {
				pa.verbose("Error %q while processing cookie, discarded\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("cookie monster ruined your day"))
				return
			}
		}
		user := pa.validate(cookie.Value)
		if user == "" {
			pa.verbose("Validation failed for token %q\n", cookie.Value)
			pa.SendAuthForm(w)
			return
		}
		pa.verbose("Token %q matches user %q\n", cookie.Value, user)
		ctx := context.WithValue(r.Context(), ContextKey("username"), user)
		r = r.WithContext(ctx)
		next(w, r)
	}
}

func (pa PlaceholderAuth) createCookie(token string, lifetime Lifetime) *http.Cookie {
	return &http.Cookie{
		Name:     pa.CookieName,
		Value:    token,
		MaxAge:   int(lifetime.Seconds()),
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
	}
}

func (pa PlaceholderAuth) validateHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	returnPath := r.Referer()
	if returnPath == "" {
		returnPath = "/"
	}
	token, err := pa.authenticate(username, password)
	if err != nil {
		pa.verbose("Problem during username/password authentication: %s\n", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	if token == "" {
		pa.verbose("Could not authenticate user %q\n", username)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("credential pair presented matches no known user"))
		return
	}
	pa.verbose("User %q authenticated\n", username)
	http.SetCookie(w, pa.createCookie(token, pa.CookieLifetime))
	http.Redirect(w, r, returnPath, http.StatusSeeOther)
}

func (pa PlaceholderAuth) voidHandler(w http.ResponseWriter, r *http.Request) {
	returnPath := r.Referer()
	if returnPath == "" {
		returnPath = "/"
	}
	pa.verbose("Cookie voiding requested\n")
	http.SetCookie(w, pa.createCookie("", Lifetime{time.Second * -5}))
	http.Redirect(w, r, returnPath, http.StatusSeeOther)
}

func (pa PlaceholderAuth) authenticate(username, password string) (string, error) {
	for _, user := range pa.Users {
		if username != user.Username {
			continue
		}
		if user.PasswordHash == "" {
			pa.verbose("user %q has no password_hash, and could not be checked")
			continue
		}
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
			if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
				pa.verbose("authenticating user is not %s", user.Username)
				return "", nil
			}
			return "", fmt.Errorf("error comparing %s's password to the stored one (%s): %w", user.Username, user.PasswordHash, err)
		}

		// If we got here without a continue, it means it's the right user.
		pa.verbose("authenticated user %s", user.Username)
		return user.Token, nil
	}

	// Checked all the users, none of them matched.
	return "", nil

}

func (pa PlaceholderAuth) validate(token string) string {
	for _, user := range pa.Users {
		if user.Token == token {
			return user.Username
		}
	}
	return ""
}
