package main

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

const httpPort = 80

//go:embed index.html
var indexHTML []byte

//go:embed index.css
var indexCSS []byte

//go:embed index.js
var indexJS []byte

//go:embed alpine.min.js
var alpineJS []byte

//go:embed sleep.ipxe
var sleepIPXE []byte

//go:embed boot.ipxe
var bootIPXE []byte

func startHTTPServer() {
	http.HandleFunc("GET /", logging(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		_, err := w.Write(indexHTML)
		if err != nil {
			slog.Warn("Failed to server index.html", "err", err)
		}
	}))
	http.HandleFunc("GET /index.css", logging(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		_, err := w.Write(indexCSS)
		if err != nil {
			slog.Warn("Failed to server index.css", "err", err)
		}
	}))
	http.HandleFunc("GET /index.js", logging(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		_, err := w.Write(indexJS)
		if err != nil {
			slog.Warn("Failed to server index.js", "err", err)
		}
	}))
	http.HandleFunc("GET /alpine.js", logging(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		_, err := w.Write(alpineJS)
		if err != nil {
			slog.Warn("Failed to server index.js", "err", err)
		}
	}))
	http.HandleFunc("GET /boot/{uuid}", logging(handleBoot))
	http.HandleFunc("GET /machines", logging(handleGetMachines))

	// Serve static files from /var/lib/herrah at /files/
	fileServer := http.FileServer(http.Dir("/var/lib/herrah"))
	http.Handle("GET /files/", logging(http.StripPrefix("/files/", fileServer).ServeHTTP))

	http.HandleFunc("/", logging(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))

	slog.Info("starting http server", "port", httpPort)
	err := http.ListenAndServe(":"+strconv.Itoa(httpPort), nil)
	if err != nil {
		slog.Error("http error", "err", err)
	}
}

func handleBoot(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var m machine
	ok, err := store.Get("machine-"+uuid, &m)
	if err != nil {
		panic(err)
	}
	if !ok {
		mac := r.URL.Query().Get("mac")
		err = store.Set("machine-"+uuid, machine{
			Name:    "NEW " + uuid,
			UUID:    uuid,
			MAC:     mac,
			IP:      "TBD",
			Control: "fresh",
		})
		if err != nil {
			panic(err)
		}
		_, err = w.Write(sleepIPXE)
		if err != nil {
			panic(err)
		}
		return
	} else {
		if m.Control == "fresh" {
			_, err = w.Write(sleepIPXE)
			if err != nil {
				panic(err)
			}
			return
		}
		if m.Control == "init" {
			_, err = w.Write(bootIPXE)
			if err != nil {
				panic(err)
			}
			return
		}
	}
}

func handleGetMachines(w http.ResponseWriter, r *http.Request) {
	var ms map[string]machine
	err := store.GetByPrefix("machine-", &ms)
	if err != nil {
		panic(err)
	}
	j, err := json.Marshal(ms)
	if err != nil {
		panic(err)
	}
	w.Write(j)
}

func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request", "method", r.Method, "path", r.URL.Path, "addr", r.RemoteAddr)
		next(w, r)
	}
}
