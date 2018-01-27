package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type hashTest struct {
	Name   string
	Pass   string
	Answer string
}

var hashTests = []struct {
	name   string
	pass   string
	answer string
}{
	{name: "angryMonkey", pass: "angryMonkey", answer: "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
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

	for _, ht := range hashTests {
		t.Run(ht.name, func(t *testing.T) {
			res, err := http.PostForm(fmt.Sprintf("%s/hash", srv.URL), url.Values{"password": {ht.pass}})
			if err != nil {
				t.Fatalf("POST failed with: %s", err)
			}
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				t.Fatal("POST could not read body: %s", err)
			}
			if string(body) != ht.answer {
				t.Fatal("POST expected '%s' but got '%s'", ht.answer, body)
			}
		})
	}
}
