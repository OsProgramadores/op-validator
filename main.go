// op-validator: Challenge validator for osprogramadores.com
package main

import (
	//"crypto/sha1"
	"flag"
	"fmt"
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
	page Page
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
	challengeID := r.PostFormValue("challenge_id")
	username := r.PostFormValue("username")
	solution := r.PostFormValue("solution")

	fmt.Printf("[%s] [%s] [%s]\n", challengeID, username, solution)

	/*
		switch {
		case err != nil:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		default:
			count++
		}
	*/

	fmt.Fprintf(w, `{ "valid":"1", "token": "87236421487264641914" }`)
}

// trimSlash returns a copy of the string without a trailing slash.
func trimSlash(s string) string {
	if strings.HasSuffix(s, "/") {
		return s[:len(s)-1]
	}
	return s
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
	}

	log.Printf("Serving on port %d, using XML URL %q", *optPort, *optURL)

	// Handlers paths MUST end in /
	http.HandleFunc(u.EscapedPath()+"/", srv.rootHandler)
	http.HandleFunc(u.EscapedPath()+checkPath+"/", srv.checkHandler)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *optPort), nil))
}
