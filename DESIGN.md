# Event detection
The goal for cryptostalker is to detect new files on the file system that appear encrypted. The fsnotify package exposes operating systems file events in differing manners, so each event creation heuristic must be handled.

What we want to catch is not only a new file, but also when the file is closed for writing. So not just the creation of the inode, but when the entire file is written to disk.

## Linux
fsnotify records a CREATE event when a new file is created. The problem is that this event comes even before any writes. So if we begin checking if the file is random when we get the CREATE, the file may not have any data or too little to classify it as random.

So we check for a CREATE followed by a WRITE. It's simple and doesn't get us all the way to analyzing the entire file, but it seems good enough for now.

## OSX
fsnotify produces two types of events when a new file is written to the file system.

1. When a file is created for the first time on the volume
2. When a file is renamed from an existing file on the same volume

The first case produces events like "CREATE -> WRITE -> WRITE -> WRITE -> CHMOD"

The second case produces simply "CREATE"

So we check both types of even sequences in order to signal that the file should be analyzed for entropy.
