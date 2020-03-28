// +build ignore

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
	"strings"

	"github.com/reusee/rep"
	"gopkg.in/yaml.v2"
)

var (
	pt = fmt.Printf
)

func main() {
	var book rep.Book

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

	var infos []Word
	for word, info := range words {
		if strings.Contains(word, " ") {
			continue
		}
		infos = append(infos, info)
	}

	sort.Slice(infos, func(i, j int) bool {
		a := infos[i]
		b := infos[j]
		if a.Frequency != b.Frequency {
			return a.Frequency > b.Frequency
		}
		return a.Word < b.Word
	})

	for _, info := range infos {
		book.Entries = append(book.Entries, rep.Entry{
			Definitions: map[string]string{
				"word": info.Word,
				"def":  info.Pronounce + "\n" + strings.Join(info.Definitions, "\n"),
			},
		})
	}
	pt("%d entries\n", len(book.Entries))

	content, err = yaml.Marshal(book)
	if err != nil {
		panic(err)
	}
	if err := ioutil.WriteFile("new.yaml", content, 0644); err != nil {
		panic(err)
	}

}
