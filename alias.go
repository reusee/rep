package main

import (
	"fmt"

	"github.com/reusee/e/v2"
)

var (
	me     = e.Default.WithStack()
	ce, he = e.New(me)
	pt     = fmt.Printf
)
