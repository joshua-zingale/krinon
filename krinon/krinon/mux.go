package krinon

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
)

var KRINON_SESSION_COOKIE_NAME = "krinon_session"

type KrinonRoute interface {
	URL() *url.URL
	Scopes() []string
}

type KrinonRouter interface {
	// Gets a target URL for a given Path
	Route(path string) (KrinonRoute, error)
}

func NewKrinonMux(opts *KrinonMuxOptions) *http.ServeMux {
	mux := http.NewServeMux()

	privateKey, err := parseRSAPrivateKey(opts.PrivateKey)
	if err != nil {
		return nil
	}

	mux.HandleFunc("GET /login", func(w http.ResponseWriter, r *http.Request) {
		login(w, r, opts.OAuthConfig)
	})
	mux.HandleFunc("GET /logout", logout)

	mux.HandleFunc("GET /oauth2/callback", func(w http.ResponseWriter, r *http.Request) {
		oauth2Callback(w, r, opts.OAuthConfig, opts.Secret)
	})

	mux.HandleFunc("GET /.well-known/krinon-public-key", func(w http.ResponseWriter, r *http.Request) {
		getKrinonPublicKey(w, r, opts.PublicKey)
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpProxy(w, r, privateKey, opts.Secret, opts.Router)
	})
	return mux
}

func logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    KRINON_SESSION_COOKIE_NAME,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),

		HttpOnly: true,
	})

	next_url := "/"
	if next_url_query := r.URL.Query().Get("next"); next_url_query != "" {
		next_url = next_url_query
	}

	http.Redirect(w, r, next_url, http.StatusFound)
}

func login(w http.ResponseWriter, r *http.Request, oauthConfig *oauth2.Config) {

	state := generateRandomState(32)
	url := oauthConfig.AuthCodeURL(state)

	next_url := r.URL.Query().Get("next")
	if next_url == "" {
		next_url = "/"
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "krinon_oauth2_state",
		Value:    state,
		MaxAge:   600,
		HttpOnly: true,
	})

	http.SetCookie(w, &http.Cookie{
		Name:   "next_url",
		Value:  next_url,
		MaxAge: 600,
	})

	http.Redirect(w, r, url, http.StatusFound)
}

func oauth2Callback(w http.ResponseWriter, r *http.Request, oauthConfig *oauth2.Config, secret []byte) {
	received_state := r.FormValue("state")
	code := r.FormValue("code")

	state_in_cookie, err := r.Cookie("krinon_oauth2_state")
	if err != nil || state_in_cookie.Value != received_state {
		http.Error(w, "Invalid login attempt.", 400)
		return
	}

	tok, err := oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		log.Printf("%s", err)
		http.Error(w, "Invalid login attempt.", 400)
		return
	}

	client := oauthConfig.Client(r.Context(), tok)

	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		log.Printf("Could not fetch user info: %s", err)
		http.Error(w, "Invalid login attempt.", 400)
		return
	}

	var userInfo UserInformation
	err = json.NewDecoder(resp.Body).Decode(&userInfo)
	if err != nil {
		log.Printf("Could not parse user info: %s", err)
		http.Error(w, "Invalid login attempt.", 400)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userInfo.Sub,
		"name":  userInfo.Name,
		"email": userInfo.Email,
	})

	signedToken, err := token.SignedString(secret)
	if err != nil {
		log.Printf("Could not sign user info for jwt: %s", err)
		http.Error(w, "Invalid login attempt.", 400)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     KRINON_SESSION_COOKIE_NAME,
		Value:    signedToken,
		HttpOnly: true,
		Path:     "/",
	})

	next_url := "/"
	if next_url_cookie, err := r.Cookie("next_url"); err == nil {
		next_url = next_url_cookie.Value
	}

	http.Redirect(w, r, next_url, http.StatusFound)
}

func getKrinonPublicKey(w http.ResponseWriter, _ *http.Request, publicKey []byte) {
	w.Header().Add("Content-Type", "plain/text")
	w.Write(publicKey)
}

func httpProxy(w http.ResponseWriter, r *http.Request, privateKey *rsa.PrivateKey, secret []byte, router KrinonRouter) {

	route, err := router.Route(r.URL.Path)
	if err != nil {
		http.Error(w, "Internal Server Error: Invalid target URL", http.StatusInternalServerError)
		log.Printf("Failed to get proxy target URL: %v", err)
		return
	}

	var token *jwt.Token
	if session_jwt, err := r.Cookie(KRINON_SESSION_COOKIE_NAME); err == nil {
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(session_jwt.Value, claims, func(t *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil {
			log.Printf("Failed to parse JWT")
			http.Error(w, "Internal Server Error: Invalid target URL", http.StatusInternalServerError)
			return
		}

		token = jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
			"user_id":   claims["email"],
			"scope_ids": route.Scopes(),
			"aud":       route.URL().Host + r.URL.Path,
		})
	} else {
		token = jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{})
	}

	signedJwt, err := token.SignedString(privateKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %s", err), 500)
		log.Println("Could not sign JTW with private key")
		return
	}
	reverseProxy := httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			var cookiesToKeep []*http.Cookie
			cookies := pr.In.Cookies()
			for _, cookie := range cookies {
				if cookie.Name != KRINON_SESSION_COOKIE_NAME {
					cookiesToKeep = append(cookiesToKeep, cookie)
				}
			}

			pr.Out.Header.Del("Cookie")
			for _, cookie := range cookiesToKeep {
				pr.Out.AddCookie(cookie)
			}

			pr.SetURL(route.URL())
			pr.SetXForwarded()
			pr.Out.Header.Add("X-Krinon-JWT", signedJwt)
		},
	}
	reverseProxy.ServeHTTP(w, r)
}

type KrinonMuxOptions struct {
	PublicKey   []byte
	PrivateKey  []byte
	Secret      []byte
	OAuthConfig *oauth2.Config
	Router      KrinonRouter
}

func parseRSAPrivateKey(key []byte) (*rsa.PrivateKey, error) {

	block, _ := pem.Decode(key)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, fmt.Errorf("could not decode private key")
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %v", err)
		}
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("private key is not an RSA key")
	}

	return rsaKey, nil
}

func generateRandomState(n int) string {
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		panic(err)
	}
	return base64.RawURLEncoding.EncodeToString(bytes)
}

type UserInformation struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}
