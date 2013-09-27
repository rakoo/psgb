package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Callback string

type subscribeRequest struct {
	callback     Callback
	mode         string
	topic        Topic
	leaseSeconds int
}

type subscriber struct {
	callback     Callback
	topic        Topic
	lastNotified time.Time
	leaseSeconds int
}

// As specified by 0.4
type subscribeHandler struct {
	subscribeRequests chan *subscribeRequest
	subscribers       map[Topic]map[Callback]*subscriber // topic -> subscriber's callback -> subscriber
	challengeSource   *randStringMaker
}

func newSubscribeHandler() *subscribeHandler {

	sh := &subscribeHandler{
		subscribeRequests: make(chan *subscribeRequest),
		subscribers:       make(map[Topic]map[Callback]*subscriber),
		challengeSource:   newRandStringMaker(),
	}

	go sh.start()

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

	callback := Callback(r.FormValue("hub.callback"))
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

	topic := Topic(r.FormValue("hub.topic"))
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

func (sh *subscribeHandler) start() {
	go func() {
		for sr := range sh.subscribeRequests {
			<-FREE_CONNS
			go sh.confirmSubscription(sr)
		}
	}()
}

func (sh *subscribeHandler) confirmSubscription(sr *subscribeRequest) {

	challenge := sh.challengeSource.RandomString()

	var requestURI bytes.Buffer
	fmt.Fprint(&requestURI, string(sr.callback))
	fmt.Fprint(&requestURI, "?")
	fmt.Fprintf(&requestURI, "hub.mode=%s&", url.QueryEscape(sr.mode))
	fmt.Fprintf(&requestURI, "hub.topic=%s&", url.QueryEscape(string(sr.topic)))
	fmt.Fprintf(&requestURI, "hub.challenge=%s&", url.QueryEscape(challenge))
	fmt.Fprintf(&requestURI, "hub.lease_seconds=%d&", sr.leaseSeconds)

	log.Println("Confirming subscription for", requestURI.String())
	resp, err := http.Get(requestURI.String())
	FREE_CONNS <- true

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
		sh.subscribers[sr.topic] = make(map[Callback]*subscriber)
	}

	sub := &subscriber{
		callback:     Callback(sr.callback),
		topic:        Topic(sr.topic),
		lastNotified: time.Now(),
		leaseSeconds: sr.leaseSeconds,
	}
	sh.subscribers[sr.topic][sr.callback] = sub
}

func (sh *subscribeHandler) distributeToSubscribers(topic Topic) {
	for _, sub := range sh.subscribers[topic] {
		data := CONTENT_STORE.contentAfterDate(topic, sub.lastNotified)
		req, err := buildRequest(data, string(sub.callback), string(topic))
		if err != nil {
			continue
		}

		c := http.Client{}
		<-FREE_CONNS
		go doDistribute(c, req, 0)

	}

	// TODO: remove old elements (only keep last 10)
}

func doDistribute(c http.Client, req *http.Request, attempt int) {

	if attempt >= 5 {
		log.Printf("Failed to deliver to %s after 5 attempts. All hope is lost.", req.URL.String())
		FREE_CONNS <- true
		return
	}

	resp, err := c.Do(req)
	defer resp.Body.Close()
	FREE_CONNS <- true

	if err != nil {
		log.Printf("Error when distributing content to %s: %s", string(req.URL.String()), err.Error())

		// Wait 2 ^ attempt minutes before next try
		multiplier := math.Pow(2, float64(attempt))
		time.Sleep(time.Duration(int(math.Ceil(multiplier))) * time.Minute)

		<-FREE_CONNS
		go doDistribute(c, req, attempt+1)
	}
}

func buildRequest(data []byte, remoteUrl, feedUrl string) (req *http.Request, err error) {
	req, err = http.NewRequest("POST", remoteUrl, bytes.NewReader(data))
	if err != nil {
		log.Println("Couldn't create a POST request:", err.Error())
		return
	}

	var linkBuff bytes.Buffer
	fmt.Fprintf(&linkBuff, "<%s>; rel=self,", feedUrl)
	fmt.Fprintf(&linkBuff, "<%s>; rel=hub,", HUB_URL)
	req.Header.Add("Link", linkBuff.String())

	return
}
