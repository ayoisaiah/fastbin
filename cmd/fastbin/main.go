package main

import (
	"log"
	"os"

	"github.com/ayoisaiah/fastbin"
)

func main() {
	err := fastbin.Execute(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
