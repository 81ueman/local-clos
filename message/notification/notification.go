package notifiacation

import "io"

type Notification struct {
}

func New() *Notification {
	return &Notification{}
}

func (n *Notification) Marshal() ([]byte, error) {
	return nil, nil
}

func (n *Notification) UnMarshal(r io.Reader, l uint16) error {
	return nil
}
