package main

import (
	"embed"
	"encoding/json"
	"flag"
	"github.com/github/github-mcp-server/retimcp/internal/devtools"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"
)

//go:embed public/*
var assets embed.FS

func main() {
	configPath := flag.String("config", "retimcp/config/devbox.json", "Path to private operator configuration")
	flag.Parse()
	data, err := os.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}
	var config devtools.Config
	if err := json.Unmarshal(data, &config); err != nil {
		log.Fatal(err)
	}
	devHandler, err := devtools.Handler(config)
	if err != nil {
		log.Fatal(err)
	}

	target, err := url.Parse("http://127.0.0.1:8084")
	if err != nil {
		log.Fatal(err)
	}
	proxy := &httputil.ReverseProxy{Rewrite: func(r *httputil.ProxyRequest) {
		r.SetURL(target)
		r.Out.Host = target.Host
	}}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/dev" {
			devHandler.ServeHTTP(w, r)
			return
		}
		if r.URL.Path == "/.well-known/oauth-protected-resource/dev" {
			devtools.Metadata(config)(w, r)
			return
		}
		file := ""
		if r.Method == http.MethodGet || r.Method == http.MethodHead {
			if r.URL.Path == "/" && r.Header.Get("Authorization") == "" && strings.Contains(r.Header.Get("Accept"), "text/html") {
				file = "index.html"
			}
			if r.URL.Path == "/retimcp.css" {
				file = "styles.css"
			}
		}
		if file == "" {
			proxy.ServeHTTP(w, r)
			return
		}
		body, err := assets.ReadFile("public/" + file)
		if err != nil {
			http.Error(w, "Not found", 404)
			return
		}
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; style-src 'self'; base-uri 'none'; frame-ancestors 'none'")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Cache-Control", "no-cache")
		if file == "index.html" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		}
		if r.Method != http.MethodHead {
			_, _ = w.Write(body)
		}
	})
	server := &http.Server{Addr: "127.0.0.1:8082", Handler: handler, ReadHeaderTimeout: 15 * time.Second}
	log.Print("RetiMCP landing gateway listening on 127.0.0.1:8082")
	log.Fatal(server.ListenAndServe())
}
