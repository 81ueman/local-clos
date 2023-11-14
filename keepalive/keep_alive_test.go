package keepalive

import "testing"

func TestNew(t *testing.T) {
	t.Run("New", func(t *testing.T) {
		keepalive := New()
		if keepalive == nil {
			t.Errorf("keepalive is nil")
		}
	})
}

func TestMarshal(t *testing.T) {
	t.Run("Marshal", func(t *testing.T) {
		keepalive := New()
		bytes, err := keepalive.Marshal()
		if err != nil {
			t.Errorf("failed to marshal: %v", err)
		}
		if bytes != nil {
			t.Error("bytes should be nil")
		}
	})
}
