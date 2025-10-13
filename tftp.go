package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/pin/tftp/v3"
)

//go:embed ipxe.efi
var ipxeEFI []byte

//go:embed autoexec.ipxe.tmpl
var autoexecTemplate []byte

const tftpPort = 69

type tftpHook struct{}
type autoexec struct {
	Version string
}

func startTFTPServer() {
	slog.Info("Starting tftp server", "port", tftpPort)

	tftpServer := tftp.NewServer(TFTPreadHandler, nil)
	tftpServer.SetTimeout(5 * time.Second)
	tftpServer.SetHook(tftpHook{})
	err := tftpServer.ListenAndServe(":" + strconv.Itoa(tftpPort))
	if err != nil {
		slog.Error("tftp server failed", "err", err)
		os.Exit(1)
	}
	// TODO: fix shutdown
}

func TFTPreadHandler(filename string, rf io.ReaderFrom) error {
	if filename == "autoexec.ipxe" {
		tmpl, err := template.New("autoexec.ipxe").Parse(string(autoexecTemplate))
		if err != nil {
			return fmt.Errorf("parsing template: %w", err)
		}

		r, w := io.Pipe()
		errCh := make(chan error, 1)

		go func() {
			defer w.Close()
			if err := tmpl.Execute(w, autoexec{
				Version: version,
			}); err != nil {
				errCh <- fmt.Errorf("executing template: %w", err)
				return
			}
			errCh <- nil
		}()

		if _, err := rf.ReadFrom(r); err != nil {
			return fmt.Errorf("reading from pipe: %w", err)
		}

		if err := <-errCh; err != nil {
			return err
		}

		return nil
	}

	if _, err := rf.ReadFrom(bytes.NewReader(ipxeEFI)); err != nil {
		return fmt.Errorf("reading static content: %w", err)
	}

	return nil
}

func (h tftpHook) OnSuccess(stats tftp.TransferStats) {
	slog.Info("tftp success",
		"file", stats.Filename,
		"ip", stats.RemoteAddr.String(),
		"duration", stats.Duration,
	)
}

func (h tftpHook) OnFailure(stats tftp.TransferStats, err error) {
	slog.Warn("tftp fail",
		"file", stats.Filename,
		"ip", stats.RemoteAddr.String(),
		"duration", stats.Duration,
		"err", err,
	)
}
