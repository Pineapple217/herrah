package main

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"text/template"
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

//go:embed boot.ipxe.tmpl
var bootIPXETmpl []byte

//go:embed combustion.bash
var combustionBash []byte

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
	http.HandleFunc("PATCH /machines/{uuid}", logging(handlePatchMachines))
	http.HandleFunc("DELETE /machines/{uuid}", logging(handleDelMachines))
	http.HandleFunc("GET /config/{uuid}/config.ign", logging(handleGetIgnition))
	http.HandleFunc("GET /config/{uuid}/config.bash", logging(handleGetCombustion))

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
			tmpl, err := template.New("boot.ipxe").Parse(string(bootIPXETmpl))
			if err != nil {
				panic(err)
			}
			err = tmpl.Execute(w, m)
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

func handleDelMachines(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	err := store.Delete("machine-" + uuid)
	if err != nil {
		panic(err)
	}
}

func handlePatchMachines(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")

	var newm machine
	if err := json.NewDecoder(r.Body).Decode(&newm); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	slog.Info("aa edit", "m", newm)

	store.Set("machine-"+uuid, newm)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"uuid":    uuid,
		"machine": newm,
	})
}

func handleGetIgnition(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var m machine
	ok, err := store.Get("machine-"+uuid, &m)
	if err != nil {
		panic(err)
	}
	if !ok {
		slog.Error("try to fetch autoyast without first booting", "uuid", uuid)
		return
	}

	jsonData, err := GetIgnitionConfig(m)
	if err != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(jsonData)
	if err != nil {
		panic(err)
	}
}

func handleGetCombustion(w http.ResponseWriter, r *http.Request) {
	uuid := r.PathValue("uuid")
	var m machine
	ok, err := store.Get("machine-"+uuid, &m)
	if err != nil {
		panic(err)
	}
	if !ok {
		slog.Error("try to fetch autoyast without first booting", "uuid", uuid)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	_, err = w.Write(combustionBash)
	if err != nil {
		slog.Warn("Failed to server index.js", "err", err)
	}
}

func logging(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("Request", "method", r.Method, "path", r.URL.Path, "addr", r.RemoteAddr)
		next(w, r)
	}
}
