// Lightweight Link: HTTP header parsing, as per RFC5988.
// I suck at parsing, so this should work fine with PSHB and silently
// fail at params not used by PSHB

package link

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"strconv"
	"strings"
)

// Format in an HTTP header:
// Link: <http://some.host/some/path>; rel="something"
// Link: <http://some.other.host/some/path>; rel="somethingelse",<http://screw.that.header/heavily>; rel="surprise"
type Link struct {
	Uri string
	Rel string
}

func Parse(rawLink string) []*Link {
	return parseReader(bufio.NewReader(strings.NewReader(rawLink)))
}

func parseReader(reader *bufio.Reader) (ls []*Link) {
	for {
		c, err := reader.Peek(len(" "))
		if err != nil {
			return
		}
		if string(c) != " " {
			break
		}
		reader.ReadRune()
	}

	if r, _, _ := reader.ReadRune(); strconv.QuoteRune(r) != "'<'" {
		log.Println("Bad format, there should be a < to start the link, got", strconv.QuoteRune(r))
		var errBuff bytes.Buffer
		io.Copy(&errBuff, reader)
		log.Println(errBuff.String())
		return
	}

	// Read url, between < and >
	var urlBuff bytes.Buffer
	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return
		}

		if strconv.QuoteRune(r) == "'>'" {
			break
		}
		urlBuff.WriteRune(r)
	}

	// trim following spaces until ;
	for {
		r, _, _ := reader.ReadRune()
		if strconv.QuoteRune(r) == "';'" {
			break
		}
	}
	// Here, we already read ;

	rel := ""
	for {
		key, value, endOfParams, err := nextParams(reader)
		if err != nil {
			if err != io.EOF {
				log.Println("Error when parsing params:", err.Error())
				return
			}
			break
		}

		if key == "rel" {
			rel = value
		}

		if endOfParams {
			break
		}
	}

	thisLink := &Link{
		Uri: urlBuff.String(),
		Rel: rel,
	}

	allLinks := []*Link{thisLink}

	if reader.Buffered() > 0 {
		for _, l := range parseReader(reader) {
			allLinks = append(allLinks, l)
		}
	}

	return allLinks
}

func nextParams(toParse *bufio.Reader) (key, value string, endOfParams bool, err error) {

	endOfParams = false

	var paramKey bytes.Buffer
	for {
		r, _, err := toParse.ReadRune()
		if err != nil {
			return "", "", true, err
		}
		if strconv.QuoteRune(r) == "'='" {
			break
		}
		paramKey.WriteRune(r)
	}
	key = strings.Trim(paramKey.String(), "\" ")

	var paramValue bytes.Buffer
	for {
		r, _, err := toParse.ReadRune()
		if err != nil {
			if err != io.EOF {
				return "", "", true, err
			}
		}
		if strconv.QuoteRune(r) == "','" {
			endOfParams = true
		}

		if strconv.QuoteRune(r) == "';'" || strconv.QuoteRune(r) == "','" || err == io.EOF {
			break
		}

		paramValue.WriteRune(r)
	}
	value = strings.Trim(paramValue.String(), "\" ")

	return
}
