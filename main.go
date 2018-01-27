package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func HashAndEncode(pass string) string {
	hash := sha512.Sum512([]byte(pass))
	return base64.StdEncoding.EncodeToString(hash[:])
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
	fmt.Fprint(w, hash)
}

func handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/hash", hashHandler)
	return r
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <port>\n", os.Args[0])
		return
	}
	log.Fatal(http.ListenAndServe(":"+os.Args[1], handler()))
}
