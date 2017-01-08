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
	count		*int
	fastDetectPct	*int
	path		*string
	script		*string
	sleep		*int
	stopAge		*int
	window		*int
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

func isFileRandom(filename string, fastDetectPct int) bool {
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
	var t time.Time
	if fastDetectPct != 0 {
		chunkLen := 4096
		dataLen := len(data)
		chunksTotal := dataLen / chunkLen
		chunksEnc := 0
		if dataLen % chunkLen != 0 {
			chunksTotal++
		}
		t = time.Now()
		for c := 0; c < chunksTotal; c++ {
			end := c * chunkLen + chunkLen
			if end > dataLen {
				end = dataLen
			}
			if randumb.IsRandom(data[c*chunkLen:end]) {
				chunksEnc++
			}
			if chunksEnc >= chunksTotal/(100/fastDetectPct) {
				log.Println("Found suspicious file with Fast Detect")
				fmt.Println("optimized time:", time.Since(t))
				break
				//return true
			}
		}
		fmt.Println("optimized time:", time.Since(t))
	}

	t = time.Now()
	r := randumb.IsRandom(data)
	fmt.Println("unoptimized time:", time.Since(t))
	return r
}

func Stalk(opts options) {
	c := make(chan notify.EventInfo, 1)
	// Make path recursive
	rpath := filepath.Join(*opts.path, "...")
	if err := notify.Watch(rpath, c, notify.Create); err != nil {
		log.Fatal(err)
	}
	defer notify.Stop(c)

	// Ingest events forever
	for ei := range c {
		path := ei.Path()
		go func() {
			if isFileRandom(path, *opts.fastDetectPct) {
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

func flags() options {
	opts := options{
		count: flag.Int("count", 10, "The number of random files required to be seen within <window>"),
		fastDetectPct:  flag.Int("fast_detect_pct", 0, "Optimize the file analysis phase. Skip the full-file check and flag the file as random if only the first X percent of the file is random. If this check is negative, a full-file analysis will be performed in addition."),
		path:  flag.String("path", "", "The path to watch"),
		script:  flag.String("script", "", "Script to call (first parameter is the path of suspicious file) when something happens"),
		// Since the randomness check is expensive, it may make sense to sleep after
		// each check on systems that create lots of files.
		sleep:   flag.Int("sleep", 10, "The time in milliseconds to sleep before processing each new file. Adjust higher if performance is an issue."),
		stopAge: flag.Int("stopAge", 0, "Stop all processes created within the last N seconds. Default is off."),
		window:  flag.Int("window", 60, "The number of seconds within which <count> random files must be observed"),
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
