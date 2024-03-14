package main

import (
	"crypto/tls"
	"encoding/base64"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"

	"github.com/Gandalf-Le-Dev/personal-lab/abyssProxy/config"
)

var cfg *config.HTTPConfig

func init() {
	// Load configuration from file
	cfg = &config.HTTPConfig{}
	err := config.LoadConfig("./config/proxy.conf.json", cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// reverseProxy := &httputil.ReverseProxy{
	// 	Director: func(req *http.Request) {
	// 		target := cfg.Servers[req.Host]
	// 		if target.Location.ProxyPass == "" {
	// 			log.Printf("No proxy pass found for %s\n", req.Host)
	// 			return
	// 		}

	// 		if target.Location.RequiredAuth {
	// 			//TODO: Implement authentication logic here
	// 		}

	// 		println(target.Location.RequiredAuth)

	// 		req.URL.Host = target.Location.ProxyPass

	// 		if target.Location.Scheme != "" {
	// 			req.URL.Scheme = target.Location.Scheme
	// 		} else {
	// 			req.URL.Scheme = "http"
	// 		}

	// 		req.RequestURI = ""

	// 	},
	// 	Transport: http.DefaultTransport,
	// 	ErrorLog:  log.New(log.Writer(), "reverse-proxy: ", log.LstdFlags),
	// }

	// Redirect HTTP to HTTPS
	go func() {
		log.Println("HTTP server started on :80 for redirecting to HTTPS")
		log.Fatal(http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Printf("Redirecting to HTTPS from %s \n", r.Host)
			http.Redirect(w, r, "https://"+r.Host+r.RequestURI, http.StatusPermanentRedirect)
		})))
	}()

	certificate, err := tls.LoadX509KeyPair("./certs/server.pem", "./certs/server.key")
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr: ":443",
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{certificate},
			NextProtos:   []string{"h2", "http/1.1"},
		},
	}

	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reverseProxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				target := cfg.Servers[req.Host]
				if target.Location.ProxyPass == "" {
					log.Printf("No proxy pass found for %s\n", req.Host)
					return
				}

				if target.Location.RequiredAuth {
					auth := req.Header.Get("Authorization")
					username, password, ok := parseBasicAuth(auth)
					// Fetch credentials from configuration or a secure source
					correctUsername := target.Location.Auth.Username
					correctPassword := target.Location.Auth.Password

					if !ok || username != correctUsername || password != correctPassword {
						// Return hello world written in html
						// w.Header().Set("Content-Type", "text/html; charset=utf-8")
						// w.WriteHeader(http.StatusUnauthorized)
						// w.Write([]byte("<h1>Hello World</h1>"))

						w.Header().Set("WWW-Authenticate", "Basic realm=\"Restricted\"")
						w.WriteHeader(http.StatusUnauthorized)
						http.Error(w, "Unauthorized", http.StatusUnauthorized)
						return // Stop processing the request
					}
				}

				if target.Location.Scheme == "" {
					req.URL.Scheme = "http"
				} else {
					req.URL.Scheme = target.Location.Scheme
				}

				req.URL.Host = target.Location.ProxyPass
				req.RequestURI = ""

			},
			Transport: http.DefaultTransport,
			ErrorLog:  log.New(log.Writer(), "reverse-proxy: ", log.LstdFlags),
		}

		reverseProxy.ServeHTTP(w, r)
	})

	// Start HTTPS server
	log.Printf("HTTPS server started on %s\n", server.Addr)
	ln, err := tls.Listen("tcp", server.Addr, server.TLSConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Serve(ln))
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	if strings.HasPrefix(auth, "Basic ") {
		payload, err := base64.StdEncoding.DecodeString(auth[6:])
		if err == nil {
			pair := strings.SplitN(string(payload), ":", 2)
			if len(pair) == 2 {
				return pair[0], pair[1], true
			}
		}
	}
	return
}
