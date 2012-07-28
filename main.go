package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// return the href value of an a-html-element
func ExtractHref(aElem xml.StartElement) string {
	for _, attr := range aElem.Attr {
		if attr.Name.Local == "href" {
			return attr.Value
		}
	}
	return ""
}

// tune an xml decoder to be more relaxed an parse HTML
func newHTMLDecoder(r io.Reader) *xml.Decoder {
	var d = xml.NewDecoder(r)
	d.Strict = false
	d.AutoClose = xml.HTMLAutoClose
	d.Entity = xml.HTMLEntity
	return d
}

// return a channel from which al outgoing links can be read
func LinkGenerator(rawurl string) <-chan *url.URL {
	var queue chan *url.URL
	if baseurl, err := url.Parse(rawurl); err == nil {
		queue = make(chan *url.URL, 20)
		go func() {
			defer close(queue)
			if reply, err := http.Get(baseurl.String()); err == nil {
				defer reply.Body.Close()
				var d = newHTMLDecoder(reply.Body)
				for token, err := d.Token(); err == nil; token, err = d.Token() {
					if t, ok := token.(xml.StartElement); ok {
						if t.Name.Local == "a" {
							if link, err := url.Parse(ExtractHref(t)); err == nil {
								queue <- link
							}
						}
					}
				}
			}
		}()
	}
	return queue
}

func init() {
	flag.Parse()
}

func main() {
	fmt.Printf("Links extracted from %v\n", flag.Arg(0))

	for l := range LinkGenerator(flag.Arg(0)) {
		fmt.Println(l)
	}
}
