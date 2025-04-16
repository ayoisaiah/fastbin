package main

import (
	"log"
	"os"

	"github.com/ayoisaiah/fastbin"
	_ "github.com/ayoisaiah/fastbin/internal/config"
	"github.com/ayoisaiah/fastbin/store"
)

func main() {
	err := store.Init()
	if err != nil {
		log.Fatal(err)
	}

	err = fastbin.Execute(os.Stdin, os.Stdout)
	if err != nil {
		log.Fatal(err)
	}
}
