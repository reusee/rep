package main

import (
	"flag"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type Book struct {
	Entries map[string]*Entry
	Modes   [][2]string
}

type Entry struct {
	Definitions map[string]string
	Frequency   int
	Practices   []Practice
}

type Practice struct {
	Time time.Time
	Key  string // key of Entry.Definitions
	Pass bool
}

var (
	flagUpdate = flag.Bool("update", false, "update book")
)

func init() {
	flag.Parse()
}

func main() {
	args := flag.Args()
	path := args[0]

	book := new(Book)

	if content, err := ioutil.ReadFile(path); err == nil {
		ce(yaml.Unmarshal(content, &book))
	}
	pt("loaded %d entries from %s\n", len(book.Entries), path)

	if *flagUpdate {
		updateBook(book)
	}

	content, err := yaml.Marshal(book)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile(path, content, 0644); err != nil {
		panic(err)
	}

}
