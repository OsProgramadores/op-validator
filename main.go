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
	"path/filepath"
	"strings"
	"text/template"
)

const (
	// Path for the result checker handler.
	checkPath = "/check"

	// Path for the verify-token handler.
	verifyTokenPath = "/verify-token"
)

// Page holds values to be passed to the page templates.
type Page struct {
	CheckURL string

	// Completed Challenges.
	Results []Result
}

// Server holds database and other information about this server.
type Server struct {
	page         Page
	secret       string
	rootTemplate *template.Template
}

// rootHandler serves the template to the user.
func (x *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	err := x.rootTemplate.ExecuteTemplate(w, "validate.html", x.page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// checkHandler validates the incoming request (usually a XMLHttpRequest from
// the javascript served by the template) and returns a JSON struct containing
// the validation status and the token, if valid.
func (x *Server) checkHandler(w http.ResponseWriter, r *http.Request) {
	// Form data.
	challengeID := sanitize(r.PostFormValue("challenge_id"))
	username := sanitize(r.PostFormValue("username"))
	solution := customTester(challengeID,sanitize(r.PostFormValue("solution")))
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

// validateTokenHandler validates the incoming token. It returns HTTP 200 if the token is valid
// or 400 if the token is invalid.
func (x *Server) verifyTokenHandler(w http.ResponseWriter, r *http.Request) {
	// Form data.
	challengeID := sanitize(r.PostFormValue("challenge_id"))
	username := sanitize(r.PostFormValue("username"))
	token := sanitize(r.PostFormValue("token"))

	log.Printf("Token validation: Got challenge: %q, username: %q, token: %q", challengeID, username, token)
	// Basic validation of fields.
	if challengeID == "" || username == "" || token == "" {
		log.Printf("Missing parameters: Challenge: %q, user: %q, token: %q", challengeID, username, token)
		http.Error(w, "Missing parameters", http.StatusBadRequest)
		return
	}

	// Find results for this challenge.
	result, ok := findResult(x.page.Results, challengeID)
	if !ok {
		log.Printf("Invalid challenge. Challenge: %q, user: %q", challengeID, username)
		http.Error(w, "Invalid challenge", http.StatusBadRequest)
		return
	}

	goodToken := createToken(username, x.secret, sanitize(result.Output))

	if token != goodToken {
		log.Printf("Invalid token. Got %q wanted %q", token, goodToken)
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}
	log.Printf("Good token. Challenge: %q, username: %q, token: %q", challengeID, username, token)
	http.Error(w, "OK", http.StatusOK)
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
	blanks := "\n\t\r "

	// Remove all leading & trailing spaces, tabs & newlines
	str = strings.Trim(str, blanks)
	sl := strings.Split(str, "\n")

	var ret []string
	for _, s := range sl {
		ret = append(ret, strings.Trim(s, blanks))
	}
	return strings.Join(ret, "\n")
}

// customTester executes custom tests for specific challenges
// such as those which have many solutions.
func customTester(challengeID string, solution string) string {
	if challengeID == "desafio-13" {
		if validKnightsD13(solution) {
			return "validKnightsD13"
		}
		return "notValidSolution"
	}
	return solution
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
	u, err := url.Parse(config.BaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize Template.
	rootTemplate, err := template.ParseFiles(filepath.Join(config.TemplatesDir, "validate.html"))
	if err != nil {
		log.Fatal(err)
	}

	// Create a new server object with the page parameters.
	srv := &Server{
		page: Page{
			// This is the full location where XMLHttpRequests will be sent.
			CheckURL: fmt.Sprintf("%s%s/", config.BaseURL, checkPath),
			Results:  config.Results,
		},
		secret:       config.Secret,
		rootTemplate: rootTemplate,
	}

	log.Printf("Serving on port %d", config.Port)
	log.Printf("Javascript will use XML URL: %s", config.BaseURL)

	// Handlers paths MUST end in /
	http.HandleFunc(u.EscapedPath()+"/", srv.rootHandler)
	http.HandleFunc(u.EscapedPath()+checkPath+"/", srv.checkHandler)
	http.HandleFunc(u.EscapedPath()+verifyTokenPath+"/", srv.verifyTokenHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", config.Port), nil))
}
