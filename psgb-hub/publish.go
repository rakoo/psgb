package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/url"
)

type publishHandler struct {
	newContentToFetch chan Topic // topic URI to fetch
	newContent        chan Topic // topic URI added in db
}

func newPublishHandler() *publishHandler {
	ph := &publishHandler{
		newContentToFetch: make(chan Topic),
		newContent:        make(chan Topic),
	}

	go ph.start()

	return ph
}

func (p *publishHandler) start() {
	go func() {
		for contentUri := range p.newContentToFetch {
			<-FREE_CONNS
			go p.fetchContent(contentUri)
		}
	}()
}

// As specified by 0.3
func (p *publishHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		log.Println("Error when parsing POST on publish:", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	ct := r.Header.Get("Content-Type")
	if ct != "application/x-www-form-urlencoded" {
		log.Println("Bad Content-Type in request:", ct)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	mode := r.FormValue("hub.mode")
	if mode != "publish" {
		log.Printf("Bad mode, received %s, expected %s", mode, "publish")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	rawUrls := r.Form["hub.url"]
	for _, rawUrl := range rawUrls {
		parsedUrl, err := url.Parse(rawUrl)
		if err != nil {
			log.Printf("Bad url:", err.Error())
			w.WriteHeader(http.StatusBadRequest)
			continue
		}

		log.Printf("Got new content notification for %s", parsedUrl.String())
		p.newContentToFetch <- Topic(parsedUrl.String())
	}

	w.WriteHeader(http.StatusNoContent)
}

func (p *publishHandler) fetchContent(topic Topic) {
	// TODO: User-Agent, If-None-Match, If-Modified-Since
	resp, err := http.Get(string(topic))
	FREE_CONNS <- true
	defer resp.Body.Close()

	if err != nil {
		log.Printf("Error when retrieving %s: %s", string(topic), err.Error())
		return
	}

	t := resp.Header.Get("Content-Type")
	if t == "" {
		t = http.DetectContentType([]byte(resp.Header.Get("Content-Type")))
	}

	if t != "application/atom+xml" && t != "application/rss+xml" {
		log.Println("Not parsing", t)
		return
	}

	var c bytes.Buffer
	io.Copy(&c, resp.Body)
	CONTENT_STORE.processNewContent(c.Bytes(), t, topic)

	log.Println("Got new content for", string(topic))
	p.newContent <- topic
}
