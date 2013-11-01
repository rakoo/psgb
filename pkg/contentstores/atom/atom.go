package atom

import (
	"bytes"
	"encoding/xml"
	"errors"
	"log"
	"sort"
	"strings"
)

const (
  ITEMS_IN_MEMORY := 10
)

// A struct used to get all the internal content of a feed
type entriesFeed struct {
	Entries []*entry `xml:"entry"`
}

// A struct used to get all the entries in a feed as a raw string
type rawFeed struct {
	Updated string `xml:"updated,omitempty"`
	Raw     string `xml:",innerxml"`
}

// An entry in an atom feed
type entry struct {
	Content string `xml:",innerxml"`
	Updated string `xml:"updated,omitempty"`
}

type sortableEntries []*entry

func (es sortableEntries) Len() int           { return len(es) }
func (es sortableEntries) Less(i, j int) bool { return es[i].Updated < es[j].Updated }
func (es sortableEntries) Swap(i, j int)      { es[i], es[j] = es[j], es[i] }

// A header in an atom feed
type header struct {
	Content string `xml:",innerxml"`
	Updated string `xml:"updated"`
}

type topicContent struct {
	entries []*entry
	header  string
	footer  string

	lastUpdated string
}

func (tc *topicContent) addEntries(entries []*entry) {
	sort.Sort(sortableEntries(entries))

	var lastEntryDate string
	if len(tc.entries) == 0 {
		lastEntryDate = ""
	} else {
		lastEntryDate = tc.entries[len(tc.entries)-1].Updated
	}

	for _, e := range entries {
		if e.Updated > lastEntryDate {
			e.Content = "  <entry>\n    " + strings.TrimSpace(e.Content) + "\n  </entry>\n"
			tc.entries = append(tc.entries, e)
		}
	}

  // Keep only ITEMS_IN_MEMORY items
  skip := len(tc.entries) - ITEMS_IN_MEMORY
  if skip > 0 {
    tc.entries = tc.entries[skip:]
  }
}

type AtomStore struct {
	allContent map[string]*topicContent // topic -> topic content
}

func NewStore() (as *AtomStore) {
	return &AtomStore{
		allContent: make(map[string]*topicContent),
	}
}

func (as *AtomStore) HasContent(topic string) bool {
	_, ok := as.allContent[topic]
	return ok
}

func (as *AtomStore) getTopicContent(topic string) (tc *topicContent) {
	tc, exists := as.allContent[topic]
	if !exists {
		tc = &topicContent{
			entries: make([]*entry, 0),
			header:  "",
			footer:  "",
		}
		as.allContent[topic] = tc
	}

	return
}

// Adds new content to the atom manager
func (as *AtomStore) AddNewContent(topic, content string) (lastid string, err error) {
	raw := &rawFeed{}
	err = xml.Unmarshal([]byte(content), raw)
	if err != nil {
		return "", err
	}

	atomFeed := &entriesFeed{}
	err = xml.Unmarshal([]byte(content), atomFeed)
	if err != nil {
		return "", err
	}

	entriesOnly := &entriesFeed{}
	for _, e := range atomFeed.Entries {
		newEntry := &entry{Content: e.Content}
		entriesOnly.Entries = append(entriesOnly.Entries, newEntry)
	}

	rawEntries, err := xml.MarshalIndent(entriesOnly, "", "  ")
	if err != nil {
		log.Println(err.Error())
		return "", err
	}
	reparse := &rawFeed{}
	err = xml.Unmarshal([]byte(rawEntries), reparse)
	if err != nil {
		log.Println(err.Error())
		return "", err
	}

	maybeHeaderFooter := strings.Split(content, reparse.Raw)
	if len(maybeHeaderFooter) != 2 {
		return "", errors.New("Error: Expecting 1 header and 1 footer, got something else!")
	}

	tc := as.getTopicContent(topic)
	tc.lastUpdated = raw.Updated
	tc.addEntries(atomFeed.Entries)
	tc.header = strings.TrimSpace(maybeHeaderFooter[0])
	if tc.footer == "" {
		tc.footer = strings.TrimSpace(maybeHeaderFooter[1])
	}

	return tc.lastUpdated, nil
}

// Returns content after a given id.
// This id must be parseable into a time.Time, since this is what we
// use.
func (as *AtomStore) ContentAfter(topic, id string) (content, lastId string) {
	tc := as.getTopicContent(topic)
	if tc == nil {
		return
	}

	var buf bytes.Buffer
	buf.WriteString(tc.header)
	buf.WriteString("\n")

	for idx := len(tc.entries) - 1; idx >= 0; idx-- {
		e := tc.entries[idx]
		if e.Updated >= id {
			buf.WriteString(e.Content)
		}
	}

	buf.WriteString(tc.footer)

	return strings.TrimSpace(buf.String()), tc.lastUpdated
}
