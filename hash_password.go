package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	pwd := "password123"
	hashedpwd, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}

	fmt.Printf("pwd: '%s'\nhashed pwd: '%s'\n", pwd, string(hashedpwd))

	fmt.Printf("Guess the password: ")
	var guesspwd string
	_, err = fmt.Scan(&guesspwd)
	if err != nil {
		panic(err)
	}

	err = bcrypt.CompareHashAndPassword(hashedpwd, []byte(guesspwd))
	if err != nil {
		fmt.Printf("Password check failed: '%s'\n", err)
		os.Exit(1)
	}
	fmt.Printf("Password check ok.\n")
}
