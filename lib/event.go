package lib

import (
  "runtime"
  "time"

  "github.com/fsnotify/fsnotify"
)

// Supported operating systems
const (
  Linux   = runtime.GOOS == "linux"
  OSX     = runtime.GOOS == "darwin"
  Windows = runtime.GOOS == "windows"
)

type EventDecider interface {
  RecordEvent(fsnotify.Event)
  ShouldInspect() bool
  Created()       time.Time
}
type LinuxEventDecider struct {
  EventInfo
}
type OSXEventDecider struct {
  EventInfo
}
/*
type WindowsEventDecider struct {
  EventInfo
}
*/

type EventInfo struct {
  event           fsnotify.Event
  // A CREATE event has been received for this path
  created         bool
  createdAt       time.Time
  shouldInspect   bool
  // A WRITE event has been received for this path after a CREATE
  written         bool
}
func (ei *EventInfo) ShouldInspect() bool {
  return ei.shouldInspect
}
func (ei *EventInfo) SetShouldInspect(s bool) {
  ei.shouldInspect = s
}
func (ei *EventInfo) Created() time.Time {
  return ei.createdAt
}

// RecordEvent will set the path to require inspection if:
// a) the path was recently created; and
// b) the path has since received a write
func (led *LinuxEventDecider) RecordEvent(event fsnotify.Event) {
  led.event = event
  if event.Op&fsnotify.Create == fsnotify.Create {
    led.created   = true
    led.createdAt = time.Now()
  } else if event.Op&fsnotify.Write == fsnotify.Write && led.created {
    // Unset the flag and return that this constitutes a "new file" event
    led.SetShouldInspect(true)
  }
}

// RecordEvent will set the path to require inspection if:
// a) the path was recently created; or
// b) the path was recently created, has since received a write, and received
//    a CHMOD thereafter
func (led *OSXEventDecider) RecordEvent(event fsnotify.Event) {
  led.event = event
  if event.Op&fsnotify.Create == fsnotify.Create {
    led.created   = true
    led.createdAt = time.Now()
    led.SetShouldInspect(true)
  } else if event.Op&fsnotify.Write == fsnotify.Write && led.created {
    // Unset the flag and return that this constitutes a "new file" event
    led.written   = true
  } else if event.Op&fsnotify.Chmod == fsnotify.Chmod && led.created &&
      led.written {
    led.SetShouldInspect(true)
  }
}

func Decider() EventDecider {
  if Linux {
    return &LinuxEventDecider{}
  } else if OSX {
    return &OSXEventDecider{}
  } else {
    panic("Unsupported operating system!")
  }
}
