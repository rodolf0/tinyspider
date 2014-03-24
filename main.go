package main

import (
	"bloom"
	"bufio"
	"flag"
	"log"
	"net/url"
	"os"
	"runtime"
)

var (
	baseurl     = flag.String("baseurl", "", "start url")
	numcrawlers = flag.Int("crawlers", 10, "num of concurrent crawlers")
	maxqueue    = flag.Int("maxqueue", 1e+6, "max links to queue")
)

func init() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Filter(in <-chan []*url.URL) <-chan *url.URL {
	out := make(chan *url.URL, *maxqueue)
	// kickoff
	if _url, err := url.Parse(*baseurl); err != nil {
		log.Fatalf("Couldn't parse baseurl: %v\n", err)
	} else {
		out <- _url
	}

	dups := 0
	total := 0

	filter := bloom.New(1e+7)
	go func() {
		for urls := range in {
			for _, u := range urls {
				total++
				if ustr := u.String(); !filter.Has(ustr) {
					// Drop stuff if the queue is full
					select {
					case out <- u:
						filter.Add(ustr)
					default:
						total--
					}
				} else {
					dups++
				}
			}
		}
	}()

	return out
}

func main() {
	pending := make(chan []*url.URL)
	crawlqueue := Filter(pending)
	printer := make(chan string)

	for i := 0; i < *numcrawlers; i++ {
		go func() {
			for next := range crawlqueue {
				printer <- next.String()
				pending <- NewHtmlParser(next).ExtractLinks()
			}
		}()
	}

	out := bufio.NewWriter(os.Stdout)
	for urlstr := range printer {
		out.WriteString(urlstr + "\n")
	}
}
