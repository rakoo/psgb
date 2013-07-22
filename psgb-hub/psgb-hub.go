package main

import (
	"log"
	"net/http"
)

const (
	MAX_PARALLEL_OUTGOING_CONNS = 20
	CHALLENGE_SIZE              = 20
	DEFAULT_LEASE_SECONDS       = 600
	HUB_URL                     = "http://localhost:8080"
)

var (
	ACCEPTABLE_RANDOM_CHARS = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u',
		'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7',
		'8', '9'}
	FREE_CONNS    = make(chan bool, MAX_PARALLEL_OUTGOING_CONNS)
	CONTENT_STORE = newContentStore()
)

func main() {
	for i := 0; i < MAX_PARALLEL_OUTGOING_CONNS; i++ {
		FREE_CONNS <- true
	}

	subscribeHandler := newSubscribeHandler()
	publishHandler := newPublishHandler()
	startDispatcher(subscribeHandler, publishHandler)

	http.Handle("/publish", publishHandler)
	http.Handle("/subscribe", subscribeHandler)

	log.Println("Starting server...")
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}
