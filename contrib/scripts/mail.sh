#!/bin/bash

echo "$1" | mail -s "cryptostalker alert!" $mail
