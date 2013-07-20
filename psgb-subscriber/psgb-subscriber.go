package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"

  "github.com/rakoo/psgb/pkg/link"
)

var (
	onHold                   = make(map[string]bool)
	subscriptionsOnHoldMutex sync.Mutex
)

func addSubscription(subRequest string) {
	subscriptionsOnHoldMutex.Lock()
	onHold[subRequest] = true
	subscriptionsOnHoldMutex.Unlock()
}

func removeSubscription(subRequest string) {
	subscriptionsOnHoldMutex.Lock()
	delete(onHold, subRequest)
	subscriptionsOnHoldMutex.Unlock()
}

func isOnHold(uri string) (valid bool) {
	subscriptionsOnHoldMutex.Lock()
	if _, ok := onHold[uri]; ok {
		valid = true
	} else {
		valid = false
	}
	subscriptionsOnHoldMutex.Unlock()

	return
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
		log.Println("Error in parsing the args")
		return
	}

	// As specified in 0.4
	subRequest := url.Values{}
	subRequest.Set("hub.callback", "http://localhost:8081/subscribeCallback")
	subRequest.Set("hub.topic", feedUri.String())
	subRequest.Set("hub.mode", "subscribe")

	resp, err := http.PostForm("http://localhost:8080/subscribe", subRequest)
	if err != nil {
		log.Println("Error when posting form: ", err.Error())
		return
	}

	if resp.StatusCode != 202 {
		log.Println("Got an error with subscription request: ", resp.Status)
	}

	addSubscription(feedUri.String())
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
		removeSubscription(topic)

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

	removeSubscription(topic)
	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, challenge)
	return
}

func handleNewItem(w http.ResponseWriter, r *http.Request) {
  rawLinks := r.Header[http.CanonicalHeaderKey("Link")]
  if rawLinks == nil || len(rawLinks) == 1 {
    log.Println("Missing Link: headers in update")
    return
  }

  topic := ""
  hub := ""
  for _, rawLink := range rawLinks {
    for _, parsedLink := range link.Parse(rawLink) {
      if parsedLink.Uri != "" && parsedLink.Rel != "" {
        switch (parsedLink.Rel) {
        case "rel":
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
