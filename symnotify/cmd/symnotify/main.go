package main

import (
	"log"
	"os"

	"github.com/alanconway/trials/symnotify/pkg/symnotify"
)

func main() {
	w, err := symnotify.NewWatcher()
	fatal(err)
	for _, arg := range os.Args {
		w.Add(arg)
	}
	for {
		e, err := w.Event()
		fatal(err)
		log.Print(e)
	}
}

func fatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
