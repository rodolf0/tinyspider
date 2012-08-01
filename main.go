package main

import (
	"concurrent"
	"container/list"
	"flag"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"
)

var baseurl *url.URL

func init() {
	var e error
	flag.Parse()
	if baseurl, e = url.Parse(flag.Arg(0)); e != nil {
		fmt.Fprintln(os.Stderr, "usage: tinyspider <start-url>")
		os.Exit(1)
	}
}

// print all reachable domains from a start URL
func main() {
	var visited = make(map[string]bool)
	var pending = list.New()
	var pool = concurrent.MakeWorkerPool(100)
	var condvar = sync.NewCond(&sync.Mutex{})
	var lock sync.Mutex

	var uniquedomains = make(map[string]bool)

	// signal the condvar every second to be able to check if work is done
	go func() {
		for {
			time.Sleep(time.Second)
			condvar.Signal()
		}
	}()

	pending.PushBack(baseurl)

	for {

		condvar.L.Lock()
		for !(pending.Len() > 0) {
			// if the pending queue is empty and there's no workers getting links
			if pool.Running() == 0 {
				return
			}
			condvar.Wait()
		}
		var theUrl = pending.Remove(pending.Front()).(*url.URL)
		condvar.L.Unlock()

		/*fmt.Fprintf(os.Stderr, "\r%v workers running, queue size %v", pool.Running(), pending.Len())*/

		pool.Schedule(func(u ...interface{}) {
			for l := range LinkGenerator(u[0].(*url.URL)) {
				lock.Lock()
				if _, ok := visited[l.String()]; !ok {
					visited[l.String()] = true
					pending.PushBack(l)
					condvar.Signal()
					// just print the hostname
					if _, ok := uniquedomains[l.Host]; !ok {
						uniquedomains[l.Host] = true
						fmt.Println(l.Host)
					}
				}
				lock.Unlock()
			}
		}, theUrl)
	}
}
