package utils

import (
	"io"

	log "github.com/sirupsen/logrus"
)

func CloseResourceSecure(name string, c io.Closer) {
	if c == nil {
		return
	}

	err := c.Close()
	if err != nil {
		log.Errorf("Failed to close resource '%s': %v", name, c)
	}
}

type CloserFunc struct {
	Cf func() error
}

func NewCloserFunc(f func() error) CloserFunc {
	return CloserFunc{Cf: f}
}

func (cf CloserFunc) Close() error {
	return cf.Cf()
}
