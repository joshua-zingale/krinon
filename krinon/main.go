package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joshua-zingale/krinon/krinon/internal"
	"github.com/joshua-zingale/krinon/krinon/krinon"
	"golang.org/x/oauth2"
)

func main() {
	args, err := internal.ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

	router, err := internal.NewConfigurableKrinonRouter(map[string]string{
		"mod1": "http://localhost:5555/",
	})
	if err != nil {
		panic(err)
	}

	address := fmt.Sprintf("%s:%s", args.Host, args.Port)
	server := http.Server{
		Addr: address,
		Handler: krinon.NewKrinonMux(&krinon.KrinonMuxOptions{
			Router:     router,
			PublicKey:  args.PublicKey,
			PrivateKey: args.PrivateKey,
			Secret:     args.Secret,
			OAuthConfig: &oauth2.Config{
				ClientID:     args.OAuthClientId,
				ClientSecret: args.OAuthClientSecret,
				Scopes:       []string{"email", "profile"},
				RedirectURL:  fmt.Sprintf("http://%s/oauth2/callback", address),
				Endpoint: oauth2.Endpoint{
					AuthURL:  "https://accounts.google.com/o/oauth2/auth",
					TokenURL: "https://oauth2.googleapis.com/token",
				},
			},
		}),
	}

	log.Printf("Starting HTTP server at %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Web server stopped unexpectedly: %s\n", err)
		os.Exit(1)
	}
}
