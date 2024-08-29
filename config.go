package authplaceholder

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"
)

// ErrInvalidLifetime is returned if you try to unmarshal a Lifetime that is not a string, and not a float64.
// This should not happen, but hey, famous last words!
var ErrInvalidLifetime = errors.New("invalid lifetime")

// User holds the username and hashed password for users that are accepted.
type User struct {
	Username     string `json:"username"` // Case sensitive username for this user.
	PasswordHash string `json:"password"` // bcrypt hash of a password string.
	Token        string `json:"-"`        // Generated when loading, based on the username, password, and hostname of the server.
}

// PlaceholderAuth holds the state of this particular authentication placeholder.
type PlaceholderAuth struct {
	CookieName     string   `json:"cookie_name"`     // Name of the cookie that will be set when authenticated, and removed when logging out
	CookieLifetime Lifetime `json:"cookie_lifetime"` // The lifetime (Max-Age) of the cookie.
	ValidatePath   string   `json:"validate_path"`   // The request path to handle POST requests on, to authenticate users.
	VoidPath       string   `json:"void_path"`       // The request path to handle GET requests on, to void the cookie.
	AuthTitle      string   `json:"auth_title"`      // The title of the authentication request page, and the header on that page.
	Stylesheet     string   `json:"stylesheet"`      // A full path to the CSS file you want the authentication request page to have.
	Verbose        bool     `json:"verbose"`         // If a true value, there will be a lot of babbling junk in your log.
	Users          []User   `json:"users"`           // A slice of users to be admitted if the right credentials are presented.
}

// Lifetime is really just a [time.Duration] wrapped for JSON serialization
type Lifetime struct {
	time.Duration
}

// MarshalJSON turns your Lifetime into a byte slice full of JSON data.
//
// Returns the result of the json.Marshal call, error and all.
func (lt Lifetime) MarshalJSON() ([]byte, error) {
	return json.Marshal(lt.String())
}

// UnmarshalJSON turns your byte slice full of JSON data into a Lifetime object.
//
// The error returned will most likely be because the unit given is unknown to [time.ParseDuration]
func (lt *Lifetime) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}
	switch value := v.(type) {
	case float64:
		lt.Duration = time.Duration(value)
	case string:
		dur, err := time.ParseDuration(value)
		if err != nil {
			return fmt.Errorf("unmarshal lifetime: %w", err)
		}
		lt.Duration = dur
	default:
		return ErrInvalidLifetime
	}
	return nil
}

func defaultValues() PlaceholderAuth {
	return PlaceholderAuth{
		CookieName:     "authentimication",
		CookieLifetime: Lifetime{Duration: time.Hour * 336}, // 336 hours is ~14 days
		ValidatePath:   "/auth",
		VoidPath:       "/logout",
		AuthTitle:      "Please identify yourself",
		Stylesheet:     "",
		Users:          []User{},
	}
}

func loadConfiguration(filename string) (PlaceholderAuth, error) {
	pa := defaultValues()
	file, err := os.Open(filename)
	if err != nil {
		return pa, err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&pa)
	if err != nil {
		return pa, err
	}
	host, err := os.Hostname()
	if err != nil {
		return pa, err
	}
	for i, user := range pa.Users {
		if user.PasswordHash == "" {
			return pa, fmt.Errorf("user %s has no password set", user.Username)
		}
		if len(user.PasswordHash) != 60 {
			return pa, fmt.Errorf("user %s does not seem to have a valid bcrypt password", user.Username)
		}
		token := sha256.Sum256([]byte(user.Username + host + user.PasswordHash))
		user.Token = fmt.Sprintf("%x", token)
		pa.Users[i] = user
		pa.verbose("User: %s Token: %s\n", user.Username, user.Token)
	}
	return pa, err
}
