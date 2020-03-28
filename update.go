package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"strings"
)

func updateBook(book *Book) {

	type Word struct {
		Word        string
		Pronounce   string
		Frequency   int
		Definitions []string
	}

	var words map[string]Word
	content, err := ioutil.ReadFile("collinscobuild.json")
	if err != nil {
		panic(err)
	}
	if err := json.NewDecoder(bytes.NewReader(content)).Decode(&words); err != nil {
		panic(err)
	}

	for word, info := range words {
		entry, ok := book.Entries[word]
		if !ok {
			entry = new(Entry)
			book.Entries[word] = entry
		}
		entry.Definitions = map[string]string{
			"word": strings.TrimSpace(info.Word),
			"def":  info.Pronounce + "\n" + strings.Join(info.Definitions, "\n\n"),
		}
		entry.Frequency = info.Frequency
		entry.Key = word
	}

	pt("updated %d entries\n", len(book.Entries))

	book.Modes = [][2]string{
		{"word", "def"},
	}

}
