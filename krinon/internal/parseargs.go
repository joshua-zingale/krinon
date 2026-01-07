package internal

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func ParseArgs() (*ParsedArguments, error) {

	host := flag.String("host", "127.0.0.1", "the host for this web server.")
	port := flag.String("port", "8080", "the port for this web server.")

	publicKeyFilePath := flag.String("public-jwt-key", "", "the file containing the 2048-byte RS256 public key for JWT communcation.")
	privateKeyFilePath := flag.String("private-jwt-key", "", "the file containing the 2048-byte RS256 private key for JWT communcation.")
	secretFilePath := flag.String("secret", "", "the file containing a base64 encoded random string, which will be the secret key signing cookies.")

	oauthClientIdFilePath := flag.String("oauth-client-id", "", "the file containing the client id for the oauth protocol.")
	oauthClientSecretFilePath := flag.String("oauth-client-secret", "", "the file containing the client secret for the oauth protocol.")

	flag.Parse()

	var errors []string
	publicKey, err := loadSecret("public-jwt-key", *publicKeyFilePath, "2048-byte RS256 public key")
	if err != nil {
		errors = append(errors, err.Error())
	}

	privateKey, err := loadSecret("private-jwt-key", *privateKeyFilePath, "2048-byte RS256 public key")
	if err != nil {
		errors = append(errors, err.Error())
	}

	secret, err := loadSecret("secret", *secretFilePath, "Base64 encoded random string")
	if err != nil {
		errors = append(errors, err.Error())
	}

	oauthClientId, err := loadSecret("oauth-client-id", *oauthClientIdFilePath, "OAuth client ID")
	if err != nil {
		errors = append(errors, err.Error())
	}

	oauthClientSecret, err := loadSecret("oauth-client-secret", *oauthClientSecretFilePath, "OAuth client secret")
	if err != nil {
		errors = append(errors, err.Error())
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("%s", strings.Join(errors, "\n"))
	}

	return &ParsedArguments{
		Host:              *host,
		Port:              *port,
		JwtPublicKey:      publicKey,
		JwtPrivateKey:     privateKey,
		SymmetricSecret:   secret,
		OAuthClientId:     string(oauthClientId),
		OAuthClientSecret: string(oauthClientSecret),
	}, nil

}

type ParsedArguments struct {
	Host              string
	Port              string
	JwtPublicKey      []byte
	JwtPrivateKey     []byte
	SymmetricSecret   []byte
	OAuthClientId     string
	OAuthClientSecret string
}

func loadSecret(flagName, flagPath, description string) ([]byte, error) {
	envVarName := strings.ToUpper(strings.ReplaceAll(flagName, "-", "_"))

	if flagPath != "" {
		content, err := os.ReadFile(flagPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s from file '%s': %w", description, flagPath, err)
		}
		return content, nil
	}

	if envValue := os.Getenv(envVarName); envValue != "" {
		return []byte(envValue), nil
	}

	return nil, fmt.Errorf("must either specify the flag --%s (a file path) or set the environment variable %s (%s)", flagName, envVarName, description)
}
