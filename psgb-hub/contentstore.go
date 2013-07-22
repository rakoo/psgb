package main

import (
	"bytes"
	"encoding/xml"
	"github.com/rakoo/feeds"
	"log"
	"sort"
	"time"
)

type Topic string

type contentStore struct {
	contentHeader      map[Topic]string               // topic -> header
	contentSortedItems map[Topic][]time.Time          // topic -> sorted list of updated date
	content            map[Topic]map[time.Time]string // topic -> updated date -> item content
}

func newContentStore() (cs *contentStore) {
	return &contentStore{
		contentHeader:      make(map[Topic]string),
		contentSortedItems: make(map[Topic][]time.Time),
		content:            make(map[Topic]map[time.Time]string),
	}
}

func (cs *contentStore) processNewContent(rawContent []byte, ct string, topic Topic) {
	switch ct {
	case "applicatio/atom+xml":
		cs.processAtom(rawContent, topic)
	case "applicatio/rss+xml":
		cs.processRss(rawContent, topic)
	}

}

func (cs *contentStore) processAtom(rawContent []byte, topic Topic) {
	atomFeed := &feeds.AtomFeed{}
	err := xml.Unmarshal(rawContent, atomFeed)
	if err != nil {
		log.Println("Couldn't parse atom content", err.Error())
		return
	}

	items := cs.content[topic]
	if items == nil {
		cs.content[topic] = make(map[time.Time]string)
		items = cs.content[topic]
	}

	sortedDates := cs.contentSortedItems[topic]
	if sortedDates == nil {
		cs.contentSortedItems[topic] = make([]time.Time, len(atomFeed.Entries))
		sortedDates = cs.contentSortedItems[topic]
	}

	for _, newItem := range atomFeed.Entries {
		date, err := time.Parse(time.RFC3339, newItem.Updated)
		if err != nil {
			log.Printf("Couldn't parse %s as a RFC3339 date. Not accepting this.", newItem.Updated)
			continue
		}

		content, err := xml.MarshalIndent(newItem, "", "  ")
		if err != nil {
			log.Println("Couldn't re-marshal element:", err.Error())
			continue
		}

		insertDate(sortedDates, date)
		items[date] = string(content)
	}

	// since no one is supposed to use it afterwards ...
	atomFeed.Entries = []*feeds.AtomEntry{}
	header, err := xml.MarshalIndent(atomFeed, "", "  ")
	if err != nil {
		log.Println("Error when re-marshaling header:", err.Error())
	}

	cs.contentHeader[topic] = string(header)
}

func (cs *contentStore) processRss(rawContent []byte, uri Topic) {
}

func (cs *contentStore) contentAfterDate(topic Topic, t time.Time) (rawContent []byte) {
	sortedDates := cs.contentSortedItems[topic]
	searchFunc := func(i int) bool {
		return sortedDates[i].After(t) || sortedDates[i].Equal(t)
	}

	topicContent := cs.content[topic]
	b := bytes.NewBuffer(rawContent)
	for j := sort.Search(len(sortedDates), searchFunc); j < len(sortedDates); j++ {
		b.WriteString(topicContent[sortedDates[j]])
	}

	return b.Bytes()
}

func insertDate(old []time.Time, d time.Time) {
	old = append(old, d)
	sort.Sort(&dateSorter{old})
}

// A struct to sort dates by their lexicographical value. We only
// consider ISO8601 format here
type dateSorter struct {
	dates []time.Time
}

func (ds *dateSorter) Len() int {
	return len(ds.dates)
}

func (ds *dateSorter) Swap(i, j int) {
	ds.dates[i], ds.dates[j] = ds.dates[j], ds.dates[i]
}

func (ds *dateSorter) Less(i, j int) bool {
	return ds.dates[i].Before(ds.dates[j])
}
