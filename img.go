package main

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"time"
)

const microosDownload = `https://download.opensuse.org/tumbleweed/appliances/openSUSE-MicroOS.x86_64-SelfInstall.install.tar`

func DownloadToTempFile(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	tmpFile, err := os.CreateTemp("/var/lib/herrah/downloads", "herrah-download-*")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	var downloaded int64
	contentLength := resp.ContentLength

	// Wrap resp.Body to count bytes read
	reader := io.TeeReader(resp.Body, tmpFile)

	// Start progress printer
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				dl := atomic.LoadInt64(&downloaded)
				if contentLength > 0 {
					fmt.Printf("Downloaded %.2f MB / %.2f MB (%.0f%%)\n",
						float64(dl)/1024/1024,
						float64(contentLength)/1024/1024,
						float64(dl)/float64(contentLength)*100)
				} else {
					fmt.Printf("Downloaded %.2f MB\n", float64(dl)/1024/1024)
				}
			case <-stop:
				return
			}
		}
	}()

	// Copy while tracking progress
	buf := make([]byte, 32*1024)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			atomic.AddInt64(&downloaded, int64(n))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			close(stop)
			return "", err
		}
	}

	close(stop)
	fmt.Printf("Download complete: %.2f MB written to %s\n",
		float64(downloaded)/1024/1024, tmpFile.Name())

	return tmpFile.Name(), nil
}

func extractTar(tarFile, destDir string) error {
	file, err := os.Open(tarFile)
	if err != nil {
		return err
	}
	defer file.Close()

	tarReader := tar.NewReader(file)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // end of archive
		}
		if err != nil {
			return err
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		default:
			fmt.Printf("Skipping: %s\n", header.Name)
		}
	}

	return nil
}

func prefImg() error {
	path, err := DownloadToTempFile(microosDownload)
	if err != nil {
		return err
	}
	extractTar(path, "/var/lib/herrah")

	return nil
}
