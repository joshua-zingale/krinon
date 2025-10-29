package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joshua-zingale/krinon/krinon/internal"
	"github.com/joshua-zingale/krinon/krinon/krinon"
)

func main() {
	args, err := internal.ParseArgs()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Argument error: %s\n", err)
		os.Exit(1)
	}

	server := http.Server{
		Handler: krinon.NewKrinonMux(&krinon.KrinonMuxOptions{
			PublicKey:  args.PublicKey,
			PrivateKey: args.PrivateKey,
		}),
		Addr: fmt.Sprintf("%s:%s", args.Host, args.Port),
	}

	log.Printf("Starting HTTP server at %s", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("Web server stopped unexpectedly: %s\n", err)
	}
}
