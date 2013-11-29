package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
)

const (
	url = "http://10.4.31.68:8080/turn/%v"
)

func main() {
	flag.Parse()

	if flag.NArg() < 2 {
		log.Fatal("Please provide atleast 2 angles")
	}

	old := -1

	for _, as := range flag.Args() {
		a, err := strconv.Atoi(as)
		if err != nil {
			log.Fatal(err)
		}

		if old < 0 {
			old = a
			continue
		}

		if old > a {
			travel(a, old)
		} else if old < a {
			travel(old, a)
		}

		old = a
	}
}

func travel(from, to int) {
	log.Printf("Travelling from %v to %v", from, to)

	for a := from; a <= to; a++ {
		set(a)
	}
}

func set(a int) {
	u := fmt.Sprintf(url, a)
	resp, err := http.Post(u, "", nil)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	log.Printf("Set angle to %v, waiting...", a)
	time.Sleep(20 * time.Millisecond)
}
