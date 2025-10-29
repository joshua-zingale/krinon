package internal

import (
	"flag"
	"fmt"
	"os"
)

func ParseArgs() (*ParsedArguments, error) {

	host := flag.String("host", "127.0.0.1", "the host for this web server.")
	port := flag.String("port", "8080", "the port for this web server.")
	publicKeyFilePath := flag.String("public-jwt-key", "", "the file containing the 2048-byte RS256 public key for JWT communcation.")
	privateKeyFilePath := flag.String("private-jwt-key", "", "the file containing the 2048-byte RS256 private key for JWT communcation.")

	flag.Parse()

	if *publicKeyFilePath == "" {
		return nil, fmt.Errorf("must specify public-jwt-key")
	}
	if *privateKeyFilePath == "" {
		return nil, fmt.Errorf("must specify private-jwt-key")
	}

	publicKey, err := os.ReadFile(*publicKeyFilePath)
	if err != nil {
		return nil, err
	}

	privateKey, err := os.ReadFile(*privateKeyFilePath)
	if err != nil {
		return nil, err
	}

	return &ParsedArguments{
		Host:       *host,
		Port:       *port,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
	}, nil

}

type ParsedArguments struct {
	Host       string
	Port       string
	PublicKey  []byte
	PrivateKey []byte
}
