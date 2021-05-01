package watcher

type EventType int

// Compile time guarantee that the New function has been implemented for the platform.
var _ func(string) (Watcher, error) = New

const (
	EventTypeAdd EventType = iota
	EventTypeDelete
)

type Event struct {
	Path string
	Type EventType
}

type Watcher interface {
	Start()
	Stop()
	Stream() <-chan []Event
}
