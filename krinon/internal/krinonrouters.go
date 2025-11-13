package internal

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/joshua-zingale/krinon/krinon/krinon"
)

type ConfigurableKrinonRouter struct {
	pathToModuleMap map[string]*url.URL
}

type ConfiguredKrinonRoute struct {
	uRL    *url.URL
	scopes []string
}

func (r ConfiguredKrinonRoute) URL() *url.URL {
	return r.uRL
}

func (r ConfiguredKrinonRoute) Scopes() []string {
	return r.scopes
}

func NewConfigurableKrinonRouter(pathToModuleMap map[string]string) (krinon.KrinonRouter, error) {
	pathToModuleMapURL := make(map[string]*url.URL)

	for key, value := range pathToModuleMap {
		url, err := url.Parse(value)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(key, "/") {
			key = "/" + key
		}
		if !strings.HasSuffix(key, "/") {
			key += "/"
		}

		pathToModuleMapURL[key] = url
	}
	return ConfigurableKrinonRouter{
		pathToModuleMap: pathToModuleMapURL,
	}, nil
}

func (r ConfigurableKrinonRouter) Route(path string) (krinon.KrinonRoute, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if !strings.HasSuffix(path, "/") {
		path = path + "/"
	}
	for prefix, rootURL := range r.pathToModuleMap {
		if prefixMatch(prefix, path) {
			uncleanedScopes := strings.Split(prefix, "/")
			return ConfiguredKrinonRoute{
				uRL:    rootURL,
				scopes: uncleanedScopes[1 : len(uncleanedScopes)-1],
			}, nil
		}
	}
	return nil, fmt.Errorf("'%s' does not match any routes", path)
}

func prefixMatch(prefix string, possibleMatch string) bool {
	if len(prefix) > len(possibleMatch) {
		return false
	}
	return possibleMatch[:len(prefix)] == prefix
}
