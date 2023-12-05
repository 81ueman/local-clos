package notification

type Notification struct {
}

func (n *Notification) Marshal() ([]byte, error) {
	return []byte{}, nil
}
