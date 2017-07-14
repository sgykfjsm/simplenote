#!/usr/bin/env bash

set -u
set -o pipefail

url="http://localhost:8080/simplenote/api"
length=1

echo -n "Get Index: "
key="$(curl --silent "${url}/index?length=${length}" | jq --raw-output ".data[]|.key")"
echo "${key}"

echo "Get Note:"
curl --silent "${url}/get?key=${key}" | jq "."

title="$(date '+%Y%m%dT%T')"
echo -n "Create Note ${title}: "
created_key="$(echo -e "content=${title}\naaa\nbbb\nccc" | curl --silent -XPOST --data-binary @- "${url}/create" | jq --raw-output ".key")"
echo "${created_key}"
sleep 1

echo -n "Update Note ${title}: ${created_key} -> "
updated_key="$(echo -e "content=${title}\nddd\neee\nfff" | curl --silent -XPOST --data-binary @- "${url}/update?key=${created_key}" | jq --raw-output ".key")"
echo "${updated_key}"
sleep 1

echo -n "Delete Note ${title}: ${updated_key}: "
curl --silent -XPOST "${url}/delete?key=${updated_key}" | jq ".deleted"
sleep 1

echo "Delete Note Permanently ${title}"
curl --silent -XDELETE "${url}/delete?key=${updated_key}"
echo

echo "End"
