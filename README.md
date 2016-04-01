# Summary

The goal of cryptostalker is to detect crypto ransomware. The mechanism it uses to do this is by recognizing new files that are created on the file system and attempting to ascertain the likelihood that new files are encrypted.

This project is a port of the original (randumb)[github.com/unixist/randumb] project that was written in python for linux using inotify.

# Distribution
In order to be cross-platform and performant, I ported cryptostalker and the underlying randumb library to go.

(Consequently, the underlying library, fsnotify, is used for file creation notifications. The only downside compared to the old python version is that we don’t get IN_CLOSE_WRITE (Google it) behavior, which means either less performance or more narrow file creation signal. I opted for narrowness.)

### Tested on:
* Linux
* OSX
* Windows (soon)

# Usage
