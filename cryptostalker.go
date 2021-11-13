package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/rjeczalik/notify"
	"github.com/unixist/go-ps"
	"github.com/unixist/randumb"
)

type options struct {
	pathList	[]string
	count   *int
	sleep   *int
	stopAge *int
	window  *int
	script	*string
}

func stopProcsYoungerThan(secs int) {
	age, _ := time.ParseDuration(fmt.Sprintf("%ds", secs))
	for _, proc := range procsYoungerThan(age) {
		if err := stopProc(proc); err != nil {
			fmt.Printf("Failed to stop process: %d", proc)
		}
	}
}

func stopProc(pid int) error {
	if os.Getpid() == pid {
		return nil
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := p.Signal(os.Kill); err != nil {
		return err
	}
	return nil
}

func procsYoungerThan(age time.Duration) []int {
	procs, _ := ps.Processes()
	ret := []int{}
	for _, i := range procs {
		if time.Since(i.CreationTime()) < age {
			ret = append(ret, i.Pid())
		}
	}
	return ret
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
	// Processing a file in this way also has the side effect of protecting against
	// ransomware evading detection by encoding non-random data inside the file along
	// with the encrypted data--and then removing the non-random cruft data later.
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
	// Cycle each path to watch
	for _, path := range opts.pathList {
		log.Printf("Watching path: %s", path)
		
		// Make path recursive
		rpath := filepath.Join(path, "...")
		
		// Start watching path
		if err := notify.Watch(rpath, c, notify.Create); err != nil {
			log.Fatal(err)
		}
	}
	defer notify.Stop(c)

	// Ingest events forever
	for ei := range c {
		path := ei.Path()
		go func() {
			if isFileRandom(path) {
				log.Printf("Suspicious file: %s", path)
				if *opts.stopAge != 0 {
					stopProcsYoungerThan(*opts.stopAge)
				}
				if *opts.script != "" {
					exec.Command(*opts.script,path).Start()
				}
			}
		}()
		if *opts.sleep != 0 {
			time.Sleep(time.Duration(*opts.sleep) * time.Millisecond)
		}
	}
}

func usage() {
	fmt.Printf("Usage: %s [OPTIONS] path_1 [path_2] [path_N]\n", os.Args[0])
	flag.PrintDefaults()
}

func flags() options {
	// Command line switches
	opts := options{
		count: flag.Int("count", 10, "The number of random files required to be seen within <window>"),
		script:  flag.String("script", "", "Script to call (first parameter is the path of suspicious file) when something happens"),
		// Since the randomness check is expensive, it may make sense to sleep after
		// each check on systems that create lots of files.
		sleep:   flag.Int("sleep", 10, "The time in milliseconds to sleep before processing each new file. Adjust higher if performance is an issue."),
		stopAge: flag.Int("stopAge", 0, "Stop all processes created within the last N seconds. Default is off."),
		window:  flag.Int("window", 60, "The number of seconds within which <count> random files must be observed"),
	}
	
	// Usage
	flag.Usage = usage 
	flag.Parse()
	if len(os.Args) <= 1 {
		usage()
		log.Fatal("ERROR: missing one or more paths to watch as arguments")
	}
	
	// Build a list of paths to watch
	opts.pathList = []string{}
	for i:=1; i<len(os.Args);i++ {
		opts.pathList = append(opts.pathList, os.Args[i])
	}
	return opts
}

func main() {
	Stalk(flags())
}
