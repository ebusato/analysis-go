package rwi

import "gitlab.in2p3.fr/avirm/analysis-go/event"

type Reader interface {
	ReadNextEvent() (*event.Event, bool)
}
