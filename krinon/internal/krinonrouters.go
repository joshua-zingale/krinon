package internal

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/joshua-zingale/krinon/krinon/krinon"
)

type ConfigurableKrinonRouter struct {
	pathToModuleMap map[string]url.URL
}

type ConfiguredKrinonRoute struct {
	uRL      url.URL
	scopes   []string
	rootPath []string
}

func (r ConfiguredKrinonRoute) URL() url.URL {
	return r.uRL
}

func (r ConfiguredKrinonRoute) Scopes() []string {
	return r.scopes
}

func (r ConfiguredKrinonRoute) RootPath() []string {
	return r.rootPath
}

func NewConfigurableKrinonRouter(pathToModuleMap map[string]string) (krinon.KrinonRouter, error) {
	pathToModuleMapURL := make(map[string]url.URL)

	for key, value := range pathToModuleMap {
		url, err := url.Parse(value)
		if err != nil {
			return nil, err
		}
		key = normalizePath(key)

		pathToModuleMapURL[key] = *url
	}
	return ConfigurableKrinonRouter{
		pathToModuleMap: pathToModuleMapURL,
	}, nil
}

func normalizePath(path string) string {
	return strings.TrimPrefix(strings.TrimPrefix(path, "/"), "/")
}

func (r ConfigurableKrinonRouter) Route(path string) (krinon.KrinonRoute, error) {
	path = normalizePath(path)
	for prefix, rootURL := range r.pathToModuleMap {
		if prefixMatch(prefix, path) {
			prefix_list := strings.Split(prefix, "/")
			path_list := strings.Split(path, "/")
			rootURL.Path = "/" + strings.Join(strings.Split(path, "/")[len(prefix_list):], "/")

			return ConfiguredKrinonRoute{
				uRL:      rootURL,
				scopes:   path_list,
				rootPath: path_list,
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
