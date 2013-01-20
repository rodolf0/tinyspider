// crawlers are workers reading off a channel urls to explore,
// whatever edges come out of each explored url are collected
// by a deduplicator, a consumer reading off the deduplicator
// can print those urls and inject them into the explore queue

package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"time"
)

var (
	baseurl  = flag.String("baseurl", "", "start url")
	maxhosts = flag.Int("maxhosts", 0, "maximum number of hosts to print")
	// maxqueue: after maxqueue is reached urls will be output but no injected for crawling
	maxqueue = flag.Uint("maxqueue", 1<<20, "max unique urls to queue for discovery")
	startUrl *url.URL
)

func init() {
	var e error
	flag.Parse()
	if startUrl, e = url.Parse(*baseurl); e != nil {
		fmt.Fprintln(os.Stderr, "usage: tinyspider -baseurl <start-url>")
		flag.Usage()
		os.Exit(1)
	}
}

// wait on input for urls to crawl and report outgoing edges on output
func Crawler(input <-chan *url.URL) <-chan *url.URL {
	var output = make(chan *url.URL, 64)
	go func() {
		defer close(output)
		for i := range input {
			LinkGenerator(i, output)
		}
	}()
	return output
}

// fan-in input from multiple Crawlers and output unique urls
func UrlDeduplicator(inputs ...<-chan *url.URL) <-chan *url.URL {
	// emulation of select with a dynamic number of cases
	var fanin = make(chan *url.URL, len(inputs))
	var fanin_wg sync.WaitGroup
	fanin_wg.Add(len(inputs))
	for _, input_i := range inputs {
		go func(input <-chan *url.URL) {
			for i := range input {
				fanin <- i
			}
			fanin_wg.Done()
		}(input_i)
	}
	go func() {
		fanin_wg.Wait()
		close(fanin)
	}()
	var output = make(chan *url.URL)
	go func() {
		defer close(output)
		var unique_urls = make(map[string]bool)
		for i := range fanin {
			if _, ok := unique_urls[i.String()]; !ok {
				unique_urls[i.String()] = true
				output <- i
			}
		}
	}()
	return output
}

// print all reachable domains from a start URL
func main() {
	go func() {
		f, _ := os.Create("cpuprof")
		pprof.StartCPUProfile(f)
		time.Sleep(120 * time.Second)
		pprof.StopCPUProfile()
	}()

	var input = make(chan *url.URL, *maxqueue)
	var output = UrlDeduplicator(
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input),
		Crawler(input), Crawler(input))
	input <- startUrl

	var unique_hosts = make(map[string]bool)

	go func() {
		for {
			time.Sleep(10 * time.Second)
			fmt.Fprintf(os.Stderr, "go's: %v, input: %v, output: %v, hosts: %v\n",
				runtime.NumGoroutine(), len(input), len(output), len(unique_hosts))
		}
	}()

	// print unique hostnames reached from startUrl
	var buf = bufio.NewWriter(os.Stdout)
	fmt.Fprintln(buf, startUrl.Host)
	for u := range output {
		if _, ok := unique_hosts[u.Host]; !ok {
			unique_hosts[u.Host] = true
			fmt.Fprintln(buf, u.Host)
			if *maxhosts > 0 && len(unique_hosts) >= *maxhosts {
				break
			}
		}
		// don't block... we can't hold the exponential growth
		// of the web within a fixed size buffer... drop what we can't investigate
		select {
		case input <- u:
		default:
		}
	}
	close(input)
}
