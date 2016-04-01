# Summary

The goal of cryptostalker is to detect crypto ransomware. The mechanism it uses to do this is by recognizing new files that are created on the file system and trying to ascertain whether they are encrypted.

This project is a port of the original [randumb](github.com/unixist/randumb) project that was written in python for linux using inotify.

# Tested on:
* Linux
* OSX
* Windows (soon)

# Setup
These steps will set up a temporary workspace and install cryptostalker to it

#### With repo cloned

`$ source /path/to/repo/setup_workspace.sh`

#### Without repo cloned
Copy and paste these commands:

```bash
path="$HOME/workspace.$RANDOM"
export GOPATH=$path
export GOBIN=$path/bin
mkdir -p $path/src
cd $path/src
go get github.com/unixist/cryptostalker
go install github.com/unixist/cryptostalker
echo -e 'Now you can run:\n  $GOBIN/cryptostalker --path=/tmp'
```

# Example
    $./cryptostalker --path=$HOME

# Details
The file notification mechanism is Google's [fsnotify](https://github.com/fsnotify/fsnotify). Since it doesn't use the linux-specific [inotify](https://en.wikipedia.org/wiki/Inotify), cryptostalker currently relies on notifications of new files. So random/encrypted files will only be detected if they belong to new inodes. This means it wont catch the following case: a file is opened, truncated, and only then filled in with encrypted content. Fortunately, this is not how most malware works.
