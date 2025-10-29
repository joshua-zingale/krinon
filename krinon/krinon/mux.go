package krinon

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/golang-jwt/jwt"
)

func NewKrinonMux(opts *KrinonMuxOptions) *http.ServeMux {
	mux := http.NewServeMux()

	privateKey, err := parseRSAPrivateKey(opts.PrivateKey)
	if err != nil {
		return nil
	}

	mux.HandleFunc("GET /.well-known/krinon-public-key", func(w http.ResponseWriter, r *http.Request) {
		getKrinonPublicKey(opts.PublicKey, w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpProxy(privateKey, w, r)
	})
	return mux
}

func getKrinonPublicKey(publicKey []byte, w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "plain/text")
	w.Write(publicKey)
}

func httpProxy(privateKey *rsa.PrivateKey, w http.ResponseWriter, r *http.Request) {

	targetURL := "http://127.0.0.1:5555"

	url, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Internal Server Error: Invalid proxy target URL", http.StatusInternalServerError)
		log.Printf("Failed to parse proxy target URL %s: %v", targetURL, err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"user_id":  "bob@example.com",
		"scope_id": "BIO101",
	})

	signedJwt, err := token.SignedString(privateKey)
	if err != nil {
		http.Error(w, fmt.Sprintf("Internal error: %s", err), 500)
		log.Println("Could not sign JTW with private key")
		return
	}
	reverseProxy := httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			pr.SetURL(url)
			pr.SetXForwarded()
			pr.Out.Header.Add("X-Krinon-JWT", signedJwt)
		},
	}

	reverseProxy.ServeHTTP(w, r)
}

type KrinonMuxOptions struct {
	PublicKey  []byte
	PrivateKey []byte
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
