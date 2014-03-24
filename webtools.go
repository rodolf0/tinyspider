package main

import (
	"encoding/xml"
	"net/http"
	"net/url"
)

type HtmlParser struct {
	url *url.URL
}

func NewHtmlParser(baseurl interface{}) *HtmlParser {
	switch _baseurl := baseurl.(type) {
	case string:
		if _url, err := url.Parse(_baseurl); err == nil {
			return &HtmlParser{url: _url}
		}
	case *url.URL:
		if _baseurl == nil {
			return nil
		}
		return &HtmlParser{url: _baseurl}
	}
	return nil
}

func (p *HtmlParser) ExtractLinks() []*url.URL {
	var d *xml.Decoder
	if reply, err := http.Get(p.url.String()); err != nil {
		return nil
	} else {
		defer reply.Body.Close()
		d = xml.NewDecoder(reply.Body)
		d.Strict = false
		d.AutoClose = xml.HTMLAutoClose
		d.Entity = xml.HTMLEntity
	}

	extractHref := func(aElem xml.StartElement) string {
		for _, attr := range aElem.Attr {
			if attr.Name.Local == "href" {
				return attr.Value
			}
		}
		return ""
	}

	var links = make([]*url.URL, 0, 32)

	for token, err := d.Token(); err == nil; token, err = d.Token() {
		if t, ok := token.(xml.StartElement); ok && t.Name.Local == "a" {
			if href := extractHref(t); href != "" {
				if link, err := url.Parse(href); err == nil {
					link = p.url.ResolveReference(link)
					link.RawQuery = ""
					link.Fragment = ""
					links = append(links, link)
				}
			}
		}
	}

	return links
}
