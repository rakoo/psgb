package link

import (
  "testing"
)

func TestCanonicalLinkHeader(t *testing.T) {
  input := "<http://some.host/some/path>; rel=\"self\""
  links := Parse(input)

  if len(links) != 1 {
    t.Fatal("Got an unexpected number of links:", len(links))
  }

  l := links[0]

  if l.Uri == "" {
    t.Fatal("Didn't parse the uri")
  }
  if l.Uri != "http://some.host/some/path" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "http://some.host/some/path", l.Uri)
  }

  if l.Rel == "" {
    t.Fatal("Didn't parse the rel")
  }
  if l.Rel != "self" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "self", l.Rel)
  }
}

func TestCommaSeparated(t *testing.T) {
  // rel=self and rel="self" are both valid
  input := "<http://some.host/some/path>; rel=\"self\", <http://some.hub/some/path/hub>; rel=hub"
  links := Parse(input)

  if len(links) != 2 {
    t.Fatal("Got an unexpected number of links:", len(links))
  }

  selfLink := links[0]

  if selfLink.Uri == "" {
    t.Fatal("Didn't parse the uri")
  }
  if selfLink.Uri != "http://some.host/some/path" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "http://some.host/some/path", selfLink.Uri)
  }

  if selfLink.Rel == "" {
    t.Fatal("Didn't parse the rel")
  }
  if selfLink.Rel != "self" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "self", selfLink.Rel)
  }

  hubLink := links[1]
  if hubLink.Uri == "" {
    t.Fatal("Didn't parse the uri")
  }
  if hubLink.Uri != "http://some.hub/some/path/hub" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "http://some.host/some/path", hubLink.Uri)
  }

  if hubLink.Rel == "" {
    t.Fatal("Didn't parse the rel")
  }
  if hubLink.Rel != "hub" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "self", hubLink.Rel)
  }

}

func TestUnrelatedRels(t *testing.T) {
  input := "<http://some.host/some/path>; type=something; rel=\"self\"; title=\"I don't care\""
  links := Parse(input)

  if len(links) != 1 {
    t.Fatal("Got an unexpected number of links:", len(links))
  }

  l := links[0]

  if l.Uri == "" {
    t.Fatal("Didn't parse the uri")
  }
  if l.Uri != "http://some.host/some/path" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "http://some.host/some/path", l.Uri)
  }

  if l.Rel == "" {
    t.Fatal("Didn't parse the rel")
  }
  if l.Rel != "self" {
    t.Fatalf("Didn't parse the correct uri; expected %s, got %s", "self", l.Rel)
  }

}
