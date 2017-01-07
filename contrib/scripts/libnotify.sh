#!/bin/bash

## install libnotify and notify binary
##
## Debian/Ubuntu: apt-get install libnotify-bin
##

notify-send --urgency=critical "[cryptostalker] Suspicious file: $1"
