package fslist

type FSList interface {
	Pending() bool
	Add(name string) error
	Delete(name string) error
	Len() int
	Write() error
}
