// op-validator: Challenge validator for osprogramadores.com

package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"text/template"
)

const (
	// Path for the result checker handler.
	checkPath = "/check"
)

var (
	rootTemplate = template.Must(template.ParseFiles("templates/validate.html"))
)

// Page holds values to be passed to the page templates.
type Page struct {
	CheckURL string

	// Completed Challenges.
	Results []Result
}

// Server holds database and other information about this server.
type Server struct {
	// page holds information required by templates.
	page Page

	// Secret (used to compute the result hash.)
	secret string
}

// rootHandler always returns an error since we have no API endpoints here.
func (x *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	err := rootTemplate.ExecuteTemplate(w, "validate.html", x.page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// checkHandler validates the incoming request and returns a JSON
// struct containing the validation status and the token, if valid.
func (x *Server) checkHandler(w http.ResponseWriter, r *http.Request) {
	// Form data.
	challengeID := sanitize(r.PostFormValue("challenge_id"))
	username := sanitize(r.PostFormValue("username"))
	solution := sanitize(r.PostFormValue("solution"))

	log.Printf("Got challenge: %q, username: %q", challengeID, username)

	// Find corresponding result in the configuration.
	result, ok := findResult(x.page.Results, challengeID)
	if !ok {
		log.Printf("Unable to find challenge %q in config for user %q", challengeID, username)
		http.Error(w, fmt.Sprintf("Desafio inválido: %s", challengeID), http.StatusInternalServerError)
		return
	}

	good := sanitize(result.Output)

	// Log result comparison
	m := fmt.Sprintf("Username: %q, got solution %q, want %q", username, solution, good)
	if solution == good {
		m += " (GOOD)"
		fmt.Fprintf(w, `{ "valid":"1", "token": %q }`, createToken(username, x.secret, solution))
	} else {
		m += " (BAD)"
		fmt.Fprintf(w, `{ "valid":"0", "token": "" }`)
	}

	log.Print(m)
}

// findResult finds a given result in the slice of Result structures. Returns
// the result and true if found, blank and false otherwise.
func findResult(results []Result, name string) (Result, bool) {
	for _, result := range results {
		if result.Name == name {
			return result, true
		}
	}
	return Result{}, false
}

// trimSlash returns a copy of the string without a trailing slash.
func trimSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return s[:len(s)-1]
	}
	return s
}

// sanitize cleans the (possibly) multi-line string to remove common problems
// such as trailing spaces and blank lines at the beginning or end.
func sanitize(str string) string {
	// Remove all leading & trailing spaces, tabs & newlines
	str = strings.Trim(str, "\n\t ")
	sl := strings.Split(str, "\n")

	var ret []string
	for _, s := range sl {
		ret = append(ret, strings.Trim(s, "\n\t "))
	}
	return strings.Join(ret, "\n")
}

// createToken creates a token based on the username, a secret, and the result
// and returns a printable representation of the token.
func createToken(username, secret, result string) string {
	h := md5.New()
	io.WriteString(h, username)
	io.WriteString(h, secret)
	io.WriteString(h, result)

	// Always use 1 as the prefix (v1)
	return fmt.Sprintf("1%x", h.Sum(nil))
}

func main() {
	optConfig := flag.String("config", "config/op-validator.toml", "Config file name.")
	optPort := flag.Int("port", 40000, "HTTP server port to listen on.")
	optURL := flag.String("base-url", "", "Base URL for the XMLHttpRequests (from JS).")
	flag.Parse()

	// Open and parse config file.
	r, err := os.Open(*optConfig)
	if err != nil {
		log.Fatal(err)
	}
	config, err := parseConfig(r)
	if err != nil {
		log.Fatal(err)
	}

	// Remove any extra slashes from the base URL, making processing consistent.
	*optURL = trimSlash(*optURL)
	u, err := url.Parse(*optURL)
	if err != nil {
		log.Fatal(err)
	}

	// Create a new server object with the page parameters.
	srv := &Server{
		page: Page{
			// This is the full location where XMLHttpRequests will be sent.
			CheckURL: fmt.Sprintf("%s%s/", *optURL, checkPath),
			Results:  config.Results,
		},
		secret: config.Secret,
	}

	log.Printf("Serving on port %d", *optPort)
	log.Printf("Javascript will use XML URL: %s", *optURL)

	// Handlers paths MUST end in /
	http.HandleFunc(u.EscapedPath()+"/", srv.rootHandler)
	http.HandleFunc(u.EscapedPath()+checkPath+"/", srv.checkHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *optPort), nil))
}
