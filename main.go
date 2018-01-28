package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// HashAndEncode performs a sha512 hash on a string and then
// encodes the hash using base64 and returns it.
func HashAndEncode(pass string) string {
	hash := sha512.Sum512([]byte(pass))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// This will send the Interrupt signal to this process
	err := syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	if err != nil {
		log.Printf("error sending SIGINT to this process: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func hashHandler(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pass, ok := r.Form["password"]
	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if len(pass) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	hash := HashAndEncode(pass[0])
	time.Sleep(time.Until(now.Add(5 * time.Second)))
	_, err = fmt.Fprint(w, hash)
	if err != nil {
		log.Printf("error while writing hash: %s", err)
	}
}

func handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/hash", hashHandler)
	r.HandleFunc("/shutdown", shutdownHandler)
	return r
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <port>\n", os.Args[0])
		return
	}

	shutdown := make(chan os.Signal, 1)

	signal.Notify(shutdown, os.Interrupt)

	s := &http.Server{
		Addr:    ":" + os.Args[1],
		Handler: handler(),
	}

	go func() {
		log.Println("starting server")
		err := s.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("error starting http server: %s", err)
		}
	}()

	<-shutdown
	log.Println("Stopping gracefully")

	err := s.Shutdown(context.Background())
	if err != nil {
		log.Printf("error while shutting down: %s\n", err)
	}
}
