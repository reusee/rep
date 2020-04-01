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

		type Candidate struct {
			Entry       *Entry
			Proficiency time.Duration
			Spacing     time.Duration
		}
		var candidates []Candidate
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
				candidates = append(candidates, Candidate{
					Entry:       entry,
					Proficiency: 0,
					Spacing:     0,
				})
			} else if len(times) == 1 && time.Since(times[0]) > time.Hour*48 {
				candidates = append(candidates, Candidate{
					Entry:       entry,
					Proficiency: 0,
					Spacing:     time.Since(times[0]),
				})
			} else if len(times) >= 2 && time.Since(times[0]) > times[0].Sub(times[1]) {
				candidates = append(candidates, Candidate{
					Entry:       entry,
					Proficiency: times[0].Sub(times[1]),
					Spacing:     time.Since(times[0]),
				})
			}
		}
		pt("%d candidates\n", len(candidates))

		sort.Slice(candidates, func(i, j int) bool {
			a := candidates[i]
			b := candidates[j]
			if d1, d2 := a.Spacing.Round(time.Hour*18), b.Spacing.Round(time.Hour*18); d1 != d2 {
				return d1 < d2
			}
			if a.Entry.Frequency != b.Entry.Frequency {
				return a.Entry.Frequency > b.Entry.Frequency
			}
			return rand.Intn(2) == 0
		})

		if len(candidates) == 0 {
			return
		}

		n := 0
		for _, candidate := range candidates {
			a := candidate.Entry.Definitions[mode[0]]
			b := candidate.Entry.Definitions[mode[1]]
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
				candidate.Entry.Practices = append(candidate.Entry.Practices, Practice{
					Time: time.Now(),
					Mode: mode,
					Pass: true,
				})
				pt("%s\n", strings.Repeat("=", 40))
				pt("%s\n", b)
				pt("%s\n", strings.Repeat("=", 40))
				n++

			case "no", "n":
				candidate.Entry.Practices = append(candidate.Entry.Practices, Practice{
					Time: time.Now(),
					Mode: mode,
					Pass: true,
				})
				pt("%s\n", strings.Repeat("=", 40))
				pt("%s\n", b)
				pt("%s\n", strings.Repeat("=", 40))
				n++

			case "skip", "s":
				candidate.Entry.Skip = true

			case "quit", "q":
				return

			default:
				goto input

			}
			go save()
			if n > 50 {
				break
			}
		}

	}

}
