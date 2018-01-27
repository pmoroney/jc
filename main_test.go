package main

import "testing"

type hashTest struct {
	Name   string
	Pass   string
	Answer string
}

var hashTests = []hashTest{
	hashTest{
		Name:   "angryMonkey",
		Pass:   "angryMonkey",
		Answer: "ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q==",
	},
}

func TestHashAndEncode(t *testing.T) {
	for _, ht := range hashTests {
		t.Run(ht.Name, func(t *testing.T) {
			if ht.Answer != HashAndEncode(ht.Pass) {
				t.Fatalf("HashAndEncode(\"%s\") != %s", ht.Pass, ht.Answer)
			}
		})
	}
}
