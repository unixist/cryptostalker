# Summary
The goal of cryptostalker is to detect crypto ransomware. The mechanism it uses to do this is by recognizing new files that are created on the file system and trying to ascertain whether they are encrypted.

This project is a port of the original [randumb](github.com/unixist/randumb) project that was written in python for linux using inotify.

# How it works
When cryptostalker runs, it places a recursive file system watch on the path specified with the ```--path``` command line flag.

Whenever a new file is created, it is inspected for randomness via the [randumb](github.com/unixist/randumb) library. If it is deemed random, and within the ```--window``` and ```--count``` parameters, a message will be output saying that a suspicious file is found. This is possibly indicative of a newly-placed encrypted file somewhere on the filesystem under the ```--path``` directory.

If the ```--stopAge``` command line flag is specified, any new process created within ```stopAge``` seconds of an encrypted file being detected will be terminated. The idea is to stop processes that might be responsible for performing the file encryption. This is a powerful, yet dangerous feature.

Ideally, suspicious processes will be issued an interrupt, so they'd just be paused, while the user or system log is notified and you can recover any legitimate processes. Due to a limitation in golang for Windows, an interrupt can't be sent to processes; only a kill may be sent. When ```stopAge``` is implemented for other operating systems, it will be implemented with the interrupt functionality, not kill.

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
echo -e 'Now you can run:\n  $GOBIN/cryptostalker /tmp'
```

# Example
```bash
# This will print out a line if even one encrypted file is seen anywhere under $HOME
$ cryptostalker $HOME

# This will kill processes seen starting up 60 seconds before the encrypted file(s) are seen
$ cryptostalker $HOME --stopAge=60

# For performance reasons, sleep for 100 ms after checking each file for randomness
$ cryptostalker $HOME --sleep=100

# This will call a script (see contrib/scripts directory) when an encrypted file is seen anywhere under $HOME
$ cryptostalker $HOME --script=/usr/local/bin/alert.sh

# This will monitor multiple folders
$ cryptostalker /tmp $HOME
```

# Tested systems
* Linux
* OSX
* Windows

# Tested samples
* [jigsaw](https://malwr.com/analysis/MTI0NjVkYzNlMzkyNDdiZGEwZGFhZTkyNDhkMGUxZmI/)
  * Sample was detected encrypting files and terminated with the --stopAge=60
* Need more tests...

# Test your setup

## Example: GPG

### Prerequisites

* use your existing GPG key or create a new one
* cryptostalker watches a directory (e.g. ```/tmp```)


```bash
$ for i in {1..200}; do dmesg > /tmp/$i.txt; done # fill data into some files
$ for i in {1..200}; do gpg --out /tmp/$i.crypt --recipient $gpg-key-id --encrypt /tmp/$i.txt; done
```

This should result in something like:

```
YYYY/MM/DD HH:MM:SS Suspicious file: /tmp/test/70.crypt
YYYY/MM/DD HH:MM:SS Suspicious file: /tmp/test/131.crypt
YYYY/MM/DD HH:MM:SS Suspicious file: /tmp/test/165.crypt
...
```

# Details
The file notification mechanism is Google's [fsnotify](https://github.com/fsnotify/fsnotify). Since it doesn't use the linux-specific [inotify](https://en.wikipedia.org/wiki/Inotify), cryptostalker currently relies on notifications of new files. So random/encrypted files will only be detected if they belong to new inodes. This means it wont catch the following case: a file is opened, truncated, and only then filled in with encrypted content. Fortunately, this is not how most malware works.

# Bugs
There are no known bugs. But there are design choices that render the current version of cryptostalker circumventable if the malware author knows what to look for. If you're interested in discussing bypasses, we can chat directly. I'm not interested in making it easier to discover than it already is :)
