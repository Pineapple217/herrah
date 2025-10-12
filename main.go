package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
)

const version = `v0.0.0`
const banner = `
  ___ ___                             .__     
 /   |   \   _______________________  |  |__  
/    ~    \_/ __ \_  __ \_  __ \__  \ |  |  \ 
\    Y    /\  ___/|  | \/|  | \// __ \|   Y  \
 \___|_  /  \___  >__|   |__|  (____  /___|  /
       \/       \/                  \/     \/   %s

https://github.com/Pineapple217/herrah
-----------------------------------------------
`

func main() {

	if len(os.Args) < 2 {
		fmt.Println("expected 'test', 'run' or 'install' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "test":

	case "run":
		slog.SetDefault(slog.New(slog.Default().Handler()))
		fmt.Printf(banner, version)
		os.Stdout.Sync()

		go startTFTPServer()

		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("%s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
			fmt.Fprintln(w, "OK")
		})

		addr := ":80"
		log.Printf("Server listening on %s", addr)
		go func() {
			if err := http.ListenAndServe(addr, nil); err != nil {
				log.Fatal(err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		slog.Info("Received an interrupt signal, exiting...")

	case "install":

	default:
		fmt.Println("expected 'test', 'run' or 'install' subcommands")
		os.Exit(1)
	}

}
