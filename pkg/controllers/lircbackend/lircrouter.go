package lircbackend

import (
	"log"
	"path/filepath"
)

type remoteButton struct {
	remote string
	button string
}

// Handle is a function that can be registered to handle an lirc Event
type Handle func(Event)

func (l *Router) SetDefaultHandle(handle Handle) {
	l.defaultHandler = handle
}

// Handle registers a new event handler for a defined key
func (l *Router) Handle(remote string, button string, handle Handle) {
	var rb remoteButton

	if remote == "" {
		rb.remote = "*"
	} else {
		rb.remote = remote
	}

	if button == "" {
		rb.button = "*"
	} else {
		rb.button = button
	}

	if l.handlers == nil {
		l.handlers = make(map[remoteButton]Handle)
	}

	l.handlers[rb] = handle
}

// Run this in a go routine to listen for IR Key Press Events
func (l *Router) Run(killChan chan struct{}) {
	var rb remoteButton
	match := 0
	for {
		select {
		case <-killChan:
			return
		case event := <-l.receive:
			if l.defaultHandler != nil {
				l.defaultHandler(event)
			} else {
				match = 0
				// Check for exact match
				rb.remote = event.Remote
				rb.button = event.Button
				if h, ok := l.handlers[rb]; ok {
					h(event)
					continue
				}

				// Check for pattern matches
				for k, h := range l.handlers {
					remoteMatched, _ := filepath.Match(k.remote, event.Remote)
					buttonMatched, _ := filepath.Match(k.button, event.Button)

					if remoteMatched && buttonMatched {
						h(event)
						match = 1
					}
				}

				if match == 0 {
					log.Println("No match for ", event)
				}
			}
		}
	}
}
