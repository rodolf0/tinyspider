package main

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
)

// return the href value of an a-html-element
func extractHref(aElem xml.StartElement) string {
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

// fill in u with base's info when missing
func completeURL(u, base *url.URL) {
	if u.Scheme == "" {
		u.Scheme = base.Scheme
	}
	if u.Host == "" {
		u.Host = base.Host
	}
}

// return a channel from which al outgoing links can be read
func LinkGenerator(baseurl *url.URL, queue chan<- *url.URL) {
	if reply, err := http.Get(baseurl.String()); err == nil {
		defer reply.Body.Close()
		var d = newHTMLDecoder(reply.Body)
		for token, err := d.Token(); err == nil; token, err = d.Token() {
			if t, ok := token.(xml.StartElement); ok {
				if t.Name.Local == "a" {
					if link, err := url.Parse(extractHref(t)); err == nil {
						completeURL(link, baseurl)
						queue <- link
					}
				}
			}
		}
	}
}
