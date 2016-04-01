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
/*
type OSXEventDecider struct {
  EventInfo
}
type WindowsEventDecider struct {
  EventInfo
}
*/

type EventInfo struct {
  event           fsnotify.Event
  created         bool
  createdAt       time.Time
  shouldInspect   bool
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

func Decider() EventDecider {
  if Linux {
    return &LinuxEventDecider{}
  } else {
    panic("Unsupported operating system!")
  }
}
