package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/rakoo/psgb/pkg/link"
)

const (
	DEFAULT_LEASE_SECONDS = 600
)

var (
	onHold                   = make(map[string]bool)
	subscriptionsOnHoldMutex sync.Mutex
)

func addSubscriptionOnHold(topic string) {
	subscriptionsOnHoldMutex.Lock()
	defer subscriptionsOnHoldMutex.Unlock()

	onHold[topic] = true
}

func removeSubscriptionOnHold(topic string) {
	subscriptionsOnHoldMutex.Lock()
	defer subscriptionsOnHoldMutex.Unlock()

	delete(onHold, topic)
}

func isOnHold(uri string) bool {
	subscriptionsOnHoldMutex.Lock()
	defer subscriptionsOnHoldMutex.Unlock()

  _, ok := onHold[uri]
  return ok
}

func SubscribeToFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	feedUriRaw := r.FormValue("feed_uri")
	if feedUriRaw == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Didn't find feed_uri"))
		return
	}

	feedUri, err := url.Parse(feedUriRaw)
	if err != nil {
		log.Println("Error in parsing the feedUri")
		return
	}

	hubUriRaw := r.FormValue("hub_uri")
	if hubUriRaw == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Didn't find hub_uri"))
		return
	}

	hubUri, err := url.Parse(hubUriRaw)
	if err != nil {
		log.Println("Error in parsing the feedUri")
		return
	}

	addSubscriptionOnHold(feedUri.String())

	// As specified in 0.4
	subRequest := url.Values{}
	subRequest.Set("hub.callback", "http://localhost:8081/subscribeCallback")
	subRequest.Set("hub.topic", feedUri.String())
	subRequest.Set("hub.mode", "subscribe")
	subRequest.Set("hub.lease_seconds", fmt.Sprintf("%d", DEFAULT_LEASE_SECONDS))

	resp, err := http.PostForm(hubUri.String(), subRequest)
	if err != nil {
		log.Println("Error when posting form: ", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 202 {
		log.Println("Got an error with subscription request: ", resp.Status)
	}
}

func SubscribeCallbackFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		handleNewItem(w, r)
	}

	if r.Method == "GET" {
		handleVerification(w, r)
	}
}

func handleVerification(w http.ResponseWriter, r *http.Request) {
  log.Println("Confirming verification")

	err := r.ParseForm()
	if err != nil {
		log.Println("Error in parsing request: ", err.Error())
		w.WriteHeader(http.StatusOK)
		return
	}

	topic := r.FormValue("hub.topic")
	if topic == "" {
		log.Println("Couldn't find hub.topic in verification request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if !isOnHold(topic) {
		log.Println("Spammer wanted to subscribe us to ", topic)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	mode := r.FormValue("hub.mode")
	if mode == "" {
		log.Println("Couldn't find hub.mode in verification request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if mode == "denied" {
		w.WriteHeader(http.StatusOK)
		removeSubscriptionOnHold(topic)

		log.Println("Hub refused subscription to ", topic)

		reason := r.FormValue("hub.reason")
		if reason != "" {
			log.Println("\t reason: ", reason)
		}

		return
	}

	challenge := r.FormValue("hub.challenge")
	if challenge == "" {
		log.Println("Couldn't find hub.challenge in verification request")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	leaseSeconds := r.FormValue("hub.lease_seconds")

	removeSubscriptionOnHold(topic)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, challenge)

	log.Printf("Subscribed to %s for %s seconds", topic, leaseSeconds)

	return
}

func handleNewItem(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	rawLinks := r.Header[http.CanonicalHeaderKey("Link")]
	if rawLinks == nil {
		log.Println("Missing Link: headers in update")
		return
	}

	topic := ""
	hub := ""
	for _, rawLink := range rawLinks {
		for _, parsedLink := range link.Parse(rawLink) {
			if parsedLink.Uri != "" && parsedLink.Rel != "" {
				switch parsedLink.Rel {
				case "self":
					topic = parsedLink.Uri
				case "hub":
					hub = parsedLink.Uri
				}
			}
		}
	}

	log.Printf("New content for %s from %s", topic, hub)

	w.WriteHeader(http.StatusAccepted)

	return
}

func main() {
	http.HandleFunc("/subscribeTo", SubscribeToFunc)
	http.HandleFunc("/subscribeCallback", SubscribeCallbackFunc)

	log.Println("Starting subscriber on port 8081...")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
