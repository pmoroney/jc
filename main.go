package main

import (
	"context"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"expvar"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var hashCount = expvar.NewInt("hashCount")
var totalTime = expvar.NewInt("totalTime")

// hashMap is a map of hashes. The identifers are currently integers.
// If hashes were going to be saved long term, or we expect more than
// an int can store, we may want to use a string as the identifer.
// If we run are getting so many requests that we have cache contention
// for the RWMutex, we can switch to using a sync.Map instead. But that
// limits us to Go1.9+ and may be slower
type hashMap struct {
	m  map[int]string
	mu sync.RWMutex
}

// add returns the identifier for the next hash and reserves that space in the map.
// Currently it just gives the next available identifier, but if we needed
// a more opaque identifier we could pick a random one and check if it exists.
// but that adds more overhead.
func (h *hashMap) add() int {
	h.mu.Lock()
	defer h.mu.Unlock()

	i := len(h.m)
	h.m[i] = ""
	return i
}

// set provides the hashed value for the identifier.
func (h *hashMap) set(i int, hash string) {
	h.mu.Lock()
	h.m[i] = hash
	h.mu.Unlock()
}

// get returns the hash for the requested identifier.
// If the identifier does not exist, false is returned for the second return value.
func (h *hashMap) get(i int) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	hash, ok := h.m[i]
	return hash, ok
}

// hashRegistry stores hashes in memory.
// If they need to be saved between instances,
// a database or some other storage can be used.
var hashRegisty = hashMap{m: map[int]string{}}

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

func hashPostHandler(w http.ResponseWriter, r *http.Request) {
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

	id := hashRegisty.add()
	_, err = fmt.Fprintln(w, id)
	if err != nil {
		log.Printf("error while writing identifier: %s", err)
		return
	}
	f, ok := w.(http.Flusher)
	if ok {
		f.Flush()
	}

	time.Sleep(time.Until(now.Add(5 * time.Second)))

	hash := HashAndEncode(pass[0])
	hashRegisty.set(id, hash)

	_, err = fmt.Fprint(w, hash)
	if err != nil {
		log.Printf("error while writing hash: %s", err)
		return
	}
	hashCount.Add(1)
	totalTime.Add(time.Since(now).Nanoseconds())
}

func hashGetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	idString := strings.TrimPrefix(r.URL.Path, "/hash/")
	id, err := strconv.ParseInt(idString, 10, strconv.IntSize)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	hash, ok := hashRegisty.get(int(id))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if hash == "" {
		// hash hasn't been set yet. Return 404.
		// Could make it so the request waits until there is a value.
		w.WriteHeader(http.StatusNotFound)
		return
	}

	_, err = fmt.Fprint(w, hash)
	if err != nil {
		log.Printf("error while writing hash: %s", err)
		return
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	c := hashCount.Value()
	t := totalTime.Value()

	enc := json.NewEncoder(w)
	if c == 0 {
		err := enc.Encode(map[string]int64{
			"total":   0,
			"average": 0,
		})
		if err != nil {
			log.Printf("error while encoding 0 value stats: %s", err)
		}
		return
	}
	err := enc.Encode(map[string]int64{
		"total":   c,
		"average": (t / c) / 1000000,
	})
	if err != nil {
		log.Printf("error while encoding stats: %s", err)
	}
}

func handler() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/hash", hashPostHandler)
	r.HandleFunc("/hash/", hashGetHandler)
	r.HandleFunc("/shutdown", shutdownHandler)
	r.HandleFunc("/stats", statsHandler)
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
	log.Println("stopping gracefully")

	err := s.Shutdown(context.Background())
	if err != nil {
		log.Printf("error while shutting down: %s\n", err)
	}
}
