package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v2"
)

type Book struct {
	Entries map[string]*Entry
	Modes   [][2]string
}

type Entry struct {
	Key         string
	Skip        bool
	Definitions map[string]string
	Frequency   int
	Practices   []Practice
}

type Practice struct {
	Time time.Time
	Mode [2]string
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

	// load
	if content, err := ioutil.ReadFile(path); err == nil {
		ce(yaml.Unmarshal(content, &book))
	}
	pt("loaded %d entries from %s\n", len(book.Entries), path)

	// update
	if *flagUpdate {
		updateBook(book)
	}

	var saving int32
	save := func() {
		if !atomic.CompareAndSwapInt32(&saving, 0, 1) {
			return
		}
		defer atomic.CompareAndSwapInt32(&saving, 1, 0)
		content, err := yaml.Marshal(book)
		if err != nil {
			panic(err)
		}
		if err := ioutil.WriteFile(path, content, 0644); err != nil {
			panic(err)
		}
	}
	defer save()

	for _, mode := range book.Modes {

		var candidates []*Entry
		for _, entry := range book.Entries {
			if entry.Skip {
				continue
			}
			var times []time.Time
			for i := len(entry.Practices) - 1; i >= 0; i-- {
				practice := entry.Practices[i]
				if practice.Mode != mode {
					continue
				}
				if !practice.Pass {
					times = append(times, practice.Time)
					break
				}
				times = append(times, practice.Time)
				if len(times) >= 2 {
					break
				}
			}
			if len(times) == 0 {
				candidates = append(candidates, entry)
			} else if len(times) == 1 && time.Since(times[0]) > time.Hour*18 {
				candidates = append(candidates, entry)
			} else if len(times) >= 2 && time.Since(times[0]) > times[0].Sub(times[1]) {
				candidates = append(candidates, entry)
			}
		}
		pt("%d candidates\n", len(candidates))

		sort.Slice(candidates, func(i, j int) bool {
			a := candidates[i]
			b := candidates[j]
			if a.Frequency != b.Frequency {
				return a.Frequency > b.Frequency
			}
			return rand.Intn(2) == 0
		})

		if len(candidates) > 50 {
			candidates = candidates[:50]
		}

		if len(candidates) == 0 {
			return
		}

		for _, candidate := range candidates {
			a := candidate.Definitions[mode[0]]
			b := candidate.Definitions[mode[1]]
			pt("%s\n", strings.Repeat("-", 40))
			pt("%s\n", a)
			pt("%s\n", strings.Repeat("-", 40))

		input:
			var answer string
			pt("yes / no / skip -> ")
			os.Stdout.Sync()
			fmt.Scanf("%s", &answer)
			switch answer {

			case "yes", "y":
				candidate.Practices = append(candidate.Practices, Practice{
					Time: time.Now(),
					Mode: mode,
					Pass: true,
				})
				pt("%s\n", strings.Repeat("-", 40))
				pt("%s\n", b)
				pt("%s\n", strings.Repeat("-", 40))

			case "no", "n":
				candidate.Practices = append(candidate.Practices, Practice{
					Time: time.Now(),
					Mode: mode,
					Pass: true,
				})
				pt("%s\n", strings.Repeat("-", 40))
				pt("%s\n", b)
				pt("%s\n", strings.Repeat("-", 40))

			case "skip", "s":
				candidate.Skip = true

			default:
				goto input

			}
			go save()
		}

	}

}
