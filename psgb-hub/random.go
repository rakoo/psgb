package main

import (
	"bytes"
	"math/rand"
	"time"
)

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
