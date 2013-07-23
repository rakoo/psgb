#!/bin/sh

curl -XPOST -H'Content-type: application/x-www-form-urlencoded' localhost:8080/publish \
  -d 'hub.mode=publish' \
  -d 'hub.url=https://linuxfr.org/journaux.atom'
