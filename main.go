package main

import (
	"fmt"
	"log/slog"
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
------------------------------------------------------
`

var store = OpenStore("/var/lib/herrah/store.json")

func main() {

	if len(os.Args) < 2 {
		fmt.Println("expected 'test', 'run' or 'install' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "test":
		fmt.Println("TODO")
		os.Exit(0)

	case "run":
		slog.SetDefault(slog.New(slog.Default().Handler()))
		fmt.Printf(banner, version)
		os.Stdout.Sync()

		go startTFTPServer()
		go startHTTPServer()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt)
		<-quit
		slog.Info("Received an interrupt signal, exiting...")

	case "install":
		prefImg()
		os.Exit(0)

	default:
		fmt.Println("expected 'test', 'run' or 'install' subcommands")
		os.Exit(1)
	}
}
