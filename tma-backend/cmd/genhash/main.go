// genhash печатает bcrypt-хеш пароля для таблицы admins.
// Использование: go run ./cmd/genhash <пароль>
package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: genhash <password>")
		os.Exit(1)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(os.Args[1]), bcrypt.DefaultCost)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	fmt.Print(string(hash))
}
