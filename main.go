package main

import (
	"bloom"
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"time"
)

var (
	baseurl     = flag.String("baseurl", "", "start url")
	numcrawlers = flag.Int("crawlers", 10, "num of concurrent crawlers")
	maxqueue    = flag.Int("maxqueue", 1e+6, "max links to queue")
	stats       = flag.Bool("stats", true, "print stats on stdout")
)

func init() {
	flag.Parse()
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Filter(in <-chan []*url.URL, printer chan<- string) <-chan *url.URL {
	out := make(chan *url.URL, *maxqueue)
	// kickoff
	if _url, err := url.Parse(*baseurl); err != nil {
		fmt.Fprintf(os.Stderr, "Couldn't parse baseurl: %v\n", err)
		os.Exit(1)
	} else {
		out <- _url
	}

	dups := 0
	total := 0

	if *stats {
		go func() {
			tmchan := time.Tick(time.Second * 2)
			for {
				<-tmchan
				fmt.Fprintf(os.Stderr, "Total: %v, Dups: %v, qlen: %v\n", total, dups, len(out))
			}
		}()
	}

	filter := bloom.New(1e+7)
	go func() {
		for urls := range in {
			for _, u := range urls {
				total++
				if ustr := u.String(); filter.AddExisted(ustr) {
					// most probably a dup, but could be filter's false-positive
					dups++
				} else {
					printer <- u.String()
					// Drop stuff if the queue is full
					select {
					case out <- u:
					default:
						total--
					}
				}
			}
		}
	}()

	return out
}

func main() {
	pending := make(chan []*url.URL)
	printer := make(chan string)
	crawlqueue := Filter(pending, printer)

	for i := 0; i < *numcrawlers; i++ {
		go func() {
			for next := range crawlqueue {
				pending <- NewHtmlParser(next).ExtractLinks()
			}
		}()
	}

	out := bufio.NewWriter(os.Stdout)
	for urlstr := range printer {
		out.WriteString(urlstr + "\n")
	}
}
