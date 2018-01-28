package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var hashTests = []struct {
	name   string
	pass   string
	answer string
}{
	{name: "angryMonkey", pass: "angryMonkey", answer: "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
	{name: "cheese", pass: "cheese", answer: "mtBO93jfSmM/wBFk1rSFzgxsv3yr0EBeyesLN8L2XylbWkOEpZ8ebovEjGXD4uRXwSATzZHVXWIBa27r809chw=="},
}

func TestHashAndEncode(t *testing.T) {
	for _, ht := range hashTests {
		t.Run(ht.name, func(t *testing.T) {
			if ht.answer != HashAndEncode(ht.pass) {
				t.Fatalf("HashAndEncode(\"%s\") != %s", ht.pass, ht.answer)
			}
		})
	}
}

func TestRouting(t *testing.T) {
	srv := httptest.NewServer(handler())
	defer srv.Close()

	for i, ht := range hashTests {
		t.Run(ht.name, func(t *testing.T) {
			res, err := http.PostForm(fmt.Sprintf("%s/hash", srv.URL), url.Values{"password": {ht.pass}})
			if err != nil {
				t.Fatalf("POST failed with: %s", err)
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("POST could not read body: %s", err)
			}
			expected := fmt.Sprintf("%d\n%s", i, ht.answer)
			if string(body) != expected {
				t.Fatalf("POST expected '%s' but got '%s'", expected, body)
			}
		})
	}

	for i, ht := range hashTests {
		t.Run(ht.name+"-get", func(t *testing.T) {
			res, err := http.Get(fmt.Sprintf("%s/hash/%d", srv.URL, i))
			if err != nil {
				t.Fatalf("GET failed with: %s", err)
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatalf("GET could not read body: %s", err)
			}
			if string(body) != ht.answer {
				t.Fatalf("GET expected '%s' but got '%s'", ht.answer, body)
			}

		})
	}
}
