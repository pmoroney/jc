package main

import (
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"os"
)

func HashAndEncode(pass string) string {
	hash := sha512.Sum512([]byte(pass))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: %s <password>\n", os.Args[0])
		return
	}
	fmt.Print(HashAndEncode(os.Args[1]), "\n")
}
