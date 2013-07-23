#!/bin/sh

curl -XPOST localhost:8081/subscribeTo \
  -d 'feed_uri=https://linuxfr.org/journaux.atom' \
  -d 'hub_uri=http://localhost:8080/subscribe'
