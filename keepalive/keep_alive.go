package keepalive

type Keepalive struct {
}

func New() *Keepalive {
	return &Keepalive{}
}

func (k *Keepalive) Marshal() ([]byte, error) {
	return nil, nil
}
