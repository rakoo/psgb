package main

import (
	"log"

	"github.com/rakoo/psgb/pkg/contentstores/atom"
)

type formatStore interface {
	AddNewContent(topic, content string) (lastid string, err error)
	ContentAfter(topic, id string) (content, lastid string)
	HasContent(topic string) bool
}

type contentStore struct {
	formatStores map[string]formatStore
}

func newContentStore() (cs *contentStore) {
	cs = &contentStore{
		formatStores: make(map[string]formatStore),
	}

	cs.formatStores["application/atom+xml"] = atom.NewStore()

	return
}

func (cs *contentStore) processNewContent(topic, contentType, content string) {
	if fs, ok := cs.formatStores[contentType]; ok {
		_, err := fs.AddNewContent(topic, string(content))
		if err != nil {
			log.Fatal(err)
		}

	} else {
		log.Println("Couldn't parse ", contentType)
	}
}

func (cs *contentStore) contentAfter(topic, id string) (rawContent, lastid string) {
	for _, fs := range cs.formatStores {
		if fs.HasContent(topic) {
			return fs.ContentAfter(topic, id)
		}
	}

	return "", ""
}
