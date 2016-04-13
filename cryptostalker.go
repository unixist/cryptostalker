package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/rjeczalik/notify"
	"github.com/unixist/randumb"
)

type options struct {
	path *string
	count *int
	sleep *int
	window *int
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
	// if the file we're inspecting is too big. Suggestion: read data in PAGE_SIZE
	// bytes, and call randumb.IsRandom() size/PAGE_SIZE number of times. If N
	// pages are random, then return true.
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
	c := make(chan notify.EventInfo, 1)
	if err := notify.Watch(*opts.path, c, notify.Create); err != nil {
				log.Fatal(err)
	}
	defer notify.Stop(c)

	// Ingest events forever
	for ei := range c {
		path := ei.Path()
		go func() {
			if isFileRandom(path) {
				log.Printf("Suspicious file: %s", path)
			}
		}()
		time.Sleep(time.Duration(*opts.sleep) * time.Second)
	}
}

func flags() options {
	opts := options {
		count:	flag.Int("count", 10, "The number of random files required to be seen within <window>"),
		path:	 flag.String("path", "", "The path to watch"),
		// Since the randomness check is expensive, it may make sense to sleep after
		// each check on systems that create lots of files.
		sleep:	flag.Int("sleep", 1, "The time in seconds to sleep before processing each new file. Adjust higher if performance is an issue."),
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
