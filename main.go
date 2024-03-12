package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"net/http/httputil"

	"github.com/Gandalf-Le-Dev/personal-lab/abyssProxy/gen/conf"
	"golang.org/x/net/http2"
)

var cfg *conf.Proxyconf
var scheme = "http"

func init() {
	var err error
	cfg, err = conf.LoadFromPath(context.Background(), "./config/proxyconf.pkl")
	if err != nil {
		panic(err)
	}

	http2.ConfigureTransport(http.DefaultTransport.(*http.Transport))
}

func main() {
	reverseProxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			target := cfg.Http.Server[req.Host]
			if target == nil {
				target = cfg.Http.Server["localhost:443"]
			}

			req.URL.Host = target.Location.ProxyPass
			req.URL.Scheme = scheme
			req.RequestURI = ""
		},
		Transport: http.DefaultTransport,
		ErrorLog:  log.New(log.Writer(), "reverse-proxy: ", log.LstdFlags),
	}

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
		Addr:    ":443",
		Handler: reverseProxy,
		TLSConfig: &tls.Config{
			MinVersion:   tls.VersionTLS13,
			Certificates: []tls.Certificate{certificate},
			NextProtos:   []string{"h2", "http/1.1"},
		},
	}

	// Start HTTPS server
	log.Printf("HTTPS server started on %s\n", server.Addr)
	ln, err := tls.Listen("tcp", server.Addr, server.TLSConfig)
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(server.Serve(ln))
}
