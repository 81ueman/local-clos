package keepalive

import "io"

type Keepalive struct {
}

func New() *Keepalive {
	return &Keepalive{}
}

func (k *Keepalive) Marshal() ([]byte, error) {
	return nil, nil
}

func (k *Keepalive) UnMarshal(r io.Reader, l uint16) error {
	return nil
}
