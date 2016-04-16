package adapter

import "testing"

func TestBuffer(t *testing.T) {
	bufferSize := 10
	a := &Adapter{
		bufferSize: bufferSize,
	}
	buf := a.newBuffer()

	if len(buf) != 0 {
		t.Errorf("expected new buffer length to be 0 but it was %s", len(buf))
	}

	if cap(buf) != bufferSize {
		t.Errorf("expected buffer capacity to be %s but it was %s", bufferSize, cap(buf))
	}

	buf = append(buf, logglyMessage{})

	if len(buf) != 1 {
		t.Errorf(
			"expected buffer length to be 1 after adding item but it was %s",
			len(buf),
		)
	}

	oldBuffer := buf
	newBuffer := a.newBuffer()

	if &oldBuffer == &newBuffer {
		t.Error("expected oldBuffer and newBuffer to point to different structs")
	}
}
