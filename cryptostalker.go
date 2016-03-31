package main

import (
  //"flag"
  "io/ioutil"
  "log"
  "os"

  "github.com/fsnotify/fsnotify"
  "github.com/unixist/randumb"
)

func isFileRandom(filename string) bool {
  _, err := os.Stat(filename)
  if err != nil {
    log.Printf("File no longer exists: %s: %v\n", filename, err)
    return false
  }
  data, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Printf("Error reading file: %s: %v\n", filename, err)
    return false
  }
  return randumb.IsRandom(data)
}

func Stalk(path string, ) {
  watcher, err := fsnotify.NewWatcher()
  if err != nil {
    log.Fatal(err)
  }
  defer watcher.Close()

  done := make(chan bool)
  go func() {
    for {
      select {
      case event := <-watcher.Events:
        if event.Op&fsnotify.Create == fsnotify.Create {
          if isFileRandom(event.Name) {
            log.Printf("Suspicious file: %s", event.Name)
          }
        }
      case err := <-watcher.Errors:
        log.Printf("error: %v", err)
      }
    }
  }()

  err = watcher.Add(path)
  if err != nil {
    log.Fatal(err)
  }
  <-done
}

func main() {
  Stalk("/tmp")
}
