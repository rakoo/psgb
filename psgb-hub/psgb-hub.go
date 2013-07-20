package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	MAX_PARALLEL_OUTGOING_CONNS = 20
	CHALLENGE_SIZE              = 20
	DEFAULT_LEASE_SECONDS       = 600
)

var (
	ACCEPTABLE_RANDOM_CHARS = []byte{'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h',
		'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u',
		'v', 'w', 'x', 'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7',
		'8', '9'}
)

// As specified by 0.3
func publishFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello %s", r.URL.Path[1:])
}

type randStringMaker struct {
	randProv *rand.Rand
}

func newRandStringMaker() *randStringMaker {
	return &randStringMaker{
		randProv: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *randStringMaker) RandomString() string {
	var b bytes.Buffer
	for i := 0; i < CHALLENGE_SIZE; i++ {
		err := b.WriteByte(ACCEPTABLE_RANDOM_CHARS[r.randProv.Intn(len(ACCEPTABLE_RANDOM_CHARS))])
		if err != nil {
			return b.String()
		}
	}

	return b.String()
}

type subscribeRequest struct {
	callback     string
	mode         string
	topic        string
	leaseSeconds int
}

type subscriber struct {
	callback     string
	topic        string
	leaseSeconds int
}

// As specified by 0.4
type subscribeHandler struct {
	subscribeRequests chan *subscribeRequest
	subscribers       map[string]map[string]*subscriber // topic -> subscriber's callback -> subscriber
	freeConns         chan bool
	challengeSource   *randStringMaker
}

func newSubscribeHandler() *subscribeHandler {
	freeConnsChan := make(chan bool, MAX_PARALLEL_OUTGOING_CONNS)
	for i := 0; i < MAX_PARALLEL_OUTGOING_CONNS; i++ {
		freeConnsChan <- true
	}

	sh := &subscribeHandler{
		subscribeRequests: make(chan *subscribeRequest),
		subscribers:       make(map[string]map[string]*subscriber),
		challengeSource:   newRandStringMaker(),
		freeConns:         freeConnsChan,
	}

	return sh
}

func (sh *subscribeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	callback := r.FormValue("hub.callback")
	if callback == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Didn't find hub.callback"))
		return
	}

	mode := r.FormValue("hub.mode")
	if mode == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Didn't find hub.mode"))
		return
	}

	topic := r.FormValue("hub.topic")
	if topic == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Didn't find hub.topic"))
		return
	}

	leaseSecondsRaw := r.FormValue("hub.lease_seconds")
	if leaseSecondsRaw == "" {
		leaseSecondsRaw = "60"
	}
	leaseSeconds, err := strconv.Atoi(leaseSecondsRaw)
	if err != nil {
		log.Println("Error parsing %s into int, defaulting lease_seconds", leaseSecondsRaw)
		leaseSeconds = DEFAULT_LEASE_SECONDS
	}

	// TODO: hub.secret

	sh.subscribeRequests <- &subscribeRequest{
		callback:     callback,
		mode:         mode,
		topic:        topic,
		leaseSeconds: leaseSeconds,
	}

	w.WriteHeader(http.StatusAccepted)
	return
}

func (sh *subscribeHandler) confirmationLoop() {
	for sr := range sh.subscribeRequests {
		<-sh.freeConns
		go sh.confirmSubscription(sr)
	}
}

func (sh *subscribeHandler) confirmSubscription(sr *subscribeRequest) {

	challenge := sh.challengeSource.RandomString()

	var requestURI bytes.Buffer
	fmt.Fprint(&requestURI, sr.callback)
	fmt.Fprint(&requestURI, "?")
	fmt.Fprintf(&requestURI, "hub.mode=%s&", url.QueryEscape(sr.mode))
	fmt.Fprintf(&requestURI, "hub.topic=%s&", url.QueryEscape(sr.topic))
	fmt.Fprintf(&requestURI, "hub.challenge=%s&", url.QueryEscape(challenge))
	fmt.Fprintf(&requestURI, "hub.lease_seconds=%d&", sr.leaseSeconds)

	log.Println("Confirming subscription for", requestURI.String())
	resp, err := http.Get(requestURI.String())
	sh.freeConns <- true

	// TODO: put back the request in the stack to re-process it later

	if err != nil {
		log.Println("Error when confirming subscription: ", err.Error())
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		var errBuffer bytes.Buffer
		io.Copy(&errBuffer, resp.Body)
		log.Println("Error from subscriber: ", resp.Status)
		return
	}

	var bodyBuf bytes.Buffer
	io.Copy(&bodyBuf, resp.Body)
	subscriberChallenge := bodyBuf.String()

	if subscriberChallenge != challenge {
		log.Printf("Bad challenge from subscriber: expected %s, got %s", challenge, subscriberChallenge)
		return
	}

	if _, ok := sh.subscribers[sr.topic]; !ok {
		sh.subscribers[sr.topic] = make(map[string]*subscriber)
	}

	sh.subscribers[sr.topic][sr.callback] = &subscriber{
		callback:     sr.callback,
		topic:        sr.topic,
		leaseSeconds: sr.leaseSeconds,
	}

}

func main() {
	subscribeHandler := newSubscribeHandler()
	go subscribeHandler.confirmationLoop()

	http.HandleFunc("/publish", publishFunc)
	http.Handle("/subscribe", subscribeHandler)

	log.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
