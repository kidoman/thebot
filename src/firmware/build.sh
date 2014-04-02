#!/usr/bin/env bash

type go >/dev/null 2>&1 || { echo >&2 "I require go but it's not installed. Aborting."; exit 1; }
type godep >/dev/null 2>&1 || { echo >&2 "I require godep but it's not installed. Aborting."; exit 1; }

godep go build -o firmware