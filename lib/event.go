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
type GenericEventDecider struct {
  EventInfo
}

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
// a) the path was recently created; or
// b) the path was recently created, has since received a write, and received
//    a CHMOD thereafter
func (ged *GenericEventDecider) RecordEvent(event fsnotify.Event) {
  ged.event = event
  if event.Op&fsnotify.Create == fsnotify.Create {
    ged.created   = true
    ged.createdAt = time.Now()
    ged.SetShouldInspect(true)
  } else if event.Op&fsnotify.Write == fsnotify.Write && ged.created {
    // Unset the flag and return that this constitutes a "new file" event
    ged.written   = true
  } else if event.Op&fsnotify.Chmod == fsnotify.Chmod && ged.created &&
      ged.written {
    ged.SetShouldInspect(true)
  }
}

// Decider returns the appropriate EventDecider struct based upon the OS.
func Decider() EventDecider {
  if Linux || OSX || Windows {
    return &GenericEventDecider{}
  } else {
    panic("Unsupported operating system!")
  }
}
