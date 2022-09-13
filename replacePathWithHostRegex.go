package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/traefik/traefik/v2/pkg/middlewares/replacepath"
)

type Config struct {
	HostRegex       string
	PathReplacement string
}

func CreateConfig() *Config {
	return &Config{}
}

type replacePathWithHostRegex struct {
	next            http.Handler
	hostRegexp      *regexp.Regexp
	pathReplacement string
	name            string
}

func New(ctx context.Context, next http.Handler, config *Config, name string) (http.Handler, error) {
	exp, err := regexp.Compile(strings.TrimSpace(config.HostRegex))
	if err != nil {
		return nil, fmt.Errorf("error compiling regular expression %s: %w", config.HostRegex, err)
	}

	return &replacePathWithHostRegex{
		hostRegexp:      exp,
		pathReplacement: strings.TrimSpace(config.PathReplacement),
		next:            next,
		name:            name,
	}, nil
}

func (rp *replacePathWithHostRegex) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	currentPath := req.URL.RawPath
	if currentPath == "" {
		currentPath = req.URL.EscapedPath()
	}

	if rp.hostRegexp != nil && len(rp.pathReplacement) > 0 && rp.hostRegexp.MatchString(req.Host) {
		req.Header.Add(replacepath.ReplacedPathHeader, currentPath)
		req.URL.RawPath = rp.hostRegexp.ReplaceAllString(currentPath, rp.pathReplacement)

		// as replacement can introduce escaped characters
		// Path must remain an unescaped version of RawPath
		// Doesn't handle multiple times encoded replacement (`/` => `%2F` => `%252F` => ...)
		var err error
		req.URL.Path, err = url.PathUnescape(req.URL.RawPath)
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
			return
		}

		req.RequestURI = req.URL.RequestURI()
	}

	rp.next.ServeHTTP(rw, req)
}

func GetReviewAppName(host string) string {
	r := regexp.MustCompile("^app-review-([a-zA-Z0-9-]+)\\.")
	name := r.FindStringSubmatch(host)[1]
	return name

}

func main() {
	host := "app-review-fdfsd-9880ad-dasdf.dev.prowritingaid.com"
	reviewAddName := GetReviewAppName(host)

	fmt.Println(reviewAddName)
}
