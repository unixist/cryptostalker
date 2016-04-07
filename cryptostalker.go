package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/unixist/randumb"
	"github.com/unixist/cryptostalker/lib"
)

type options struct {
	path *string
	count *int
	sleep *int
	window *int
}

// Map of paths to their deciders. Includes a mutex for race-free cleanup.
type pathInfo struct {
	names map[string]lib.EventDecider
	sync.Mutex
}

func isFileRandom(filename string) bool {
	s, err := os.Stat(filename)
	if err != nil {
		// File no longer exists. Either it was a temporary file or it was removed.
		return false
	} else if !s.Mode().IsRegular() {
		// File is a directory/socket/device, anything other than a regular file.
		return false
	}
	// TODO: process the file in pieces, not as a whole. This will thrash memory
	// if the file we're inspecting is too big.
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		// Don't output an error if it is permission related
		if !os.IsPermission(err) {
			log.Printf("Error reading file: %s: %v\n", filename, err)
		}
		return false
	}
	return randumb.IsRandom(data)
}

func Stalk(opts options) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done  := make(chan bool)
	paths := pathInfo{
		names: map[string]lib.EventDecider{},
	}

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				cpath := event.Name
				_, ok := paths.names[cpath]
				if !ok {
					paths.names[cpath] = lib.Decider()
				}
				paths.Lock()
				paths.names[cpath].RecordEvent(event)
				// Decide whether the operation deems the file worthy of inspection.
				// The criteria for this vary by OS due to file system notification
				// differences.
				if paths.names[cpath].ShouldInspect() {
					go func() {
						// If required, sleep after we perform the randomness check.
						if *opts.sleep != 0.0 {
							time.Sleep(time.Duration(*opts.sleep) * time.Second)
						}
						if isFileRandom(event.Name) {
							log.Printf("Suspicious file: %s", event.Name)
						}
					}()
					// Whether it's random or not, don't inspect it again
					delete(paths.names, cpath)
				}
				paths.Unlock()
			case err := <-watcher.Errors:
				log.Printf("error: %v", err)
			}
		}
	}()

	// Run a cleanup goroutine every 10 seconds.
	// This garbage collects paths that were recorded, but never cleaned up.
	// This results in the potential for false negatives at the expense of memory
	// hygiene.
	go func() {
		for {
			paths.Lock()
			for p := range paths.names {
				if time.Since(paths.names[p].Created()) > 10 * time.Second {
					delete(paths.names, p)
				}
			}
			paths.Unlock()
			time.Sleep(10 * time.Second)
		}
	}()

	// Now with our goroutines running, begin the watch on our path and wait
	// forever.
	err = watcher.Add(*opts.path)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func flags() options {
	opts := options {
		count:  flag.Int("count", 10, "The number of random files required to be seen within <window>"),
		path:   flag.String("path", "", "The path to watch"),
		// Since the randomness check is expensive, it may make sense to sleep after
		// each check on systems that create lots of files.
		sleep:  flag.Int("sleep", 1, "The time in seconds to sleep before processing each new file. Adjust higher if performance is an issue."),
		window: flag.Int("window", 60, "The number of seconds within which <count> random files must be observed"),
	}
	flag.Parse()
	if *opts.path == "" {
		log.Fatal("Please provide a --path")
	}
	return opts
}

func main() {
	Stalk(flags())
}
