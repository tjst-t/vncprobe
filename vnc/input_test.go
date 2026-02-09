package vnc

import (
	"testing"
)

func TestSendKeySequence(t *testing.T) {
	mock := &mockClient{}
	actions := []KeyAction{
		{Key: 0xff0d, Down: true},
		{Key: 0xff0d, Down: false},
	}
	err := SendKeySequence(mock, actions)
	if err != nil {
		t.Fatalf("SendKeySequence error: %v", err)
	}
	if len(mock.keyEvents) != 2 {
		t.Fatalf("got %d key events, want 2", len(mock.keyEvents))
	}
	if mock.keyEvents[0].Key != 0xff0d || mock.keyEvents[0].Down != true {
		t.Errorf("event[0] = %+v, want {Key:0xff0d, Down:true}", mock.keyEvents[0])
	}
	if mock.keyEvents[1].Key != 0xff0d || mock.keyEvents[1].Down != false {
		t.Errorf("event[1] = %+v, want {Key:0xff0d, Down:false}", mock.keyEvents[1])
	}
}

func TestSendTypeString(t *testing.T) {
	mock := &mockClient{}
	err := SendTypeString(mock, "Hi")
	if err != nil {
		t.Fatalf("SendTypeString error: %v", err)
	}
	if len(mock.keyEvents) != 4 {
		t.Fatalf("got %d key events, want 4", len(mock.keyEvents))
	}
	wantKeys := []struct {
		key  uint32
		down bool
	}{
		{0x0048, true},
		{0x0048, false},
		{0x0069, true},
		{0x0069, false},
	}
	for i, wk := range wantKeys {
		if mock.keyEvents[i].Key != wk.key || mock.keyEvents[i].Down != wk.down {
			t.Errorf("event[%d] = {Key:0x%04x, Down:%v}, want {Key:0x%04x, Down:%v}",
				i, mock.keyEvents[i].Key, mock.keyEvents[i].Down, wk.key, wk.down)
		}
	}
}

func TestSendClick(t *testing.T) {
	mock := &mockClient{}
	err := SendClick(mock, 400, 300, 1)
	if err != nil {
		t.Fatalf("SendClick error: %v", err)
	}
	if len(mock.ptrEvents) != 2 {
		t.Fatalf("got %d pointer events, want 2", len(mock.ptrEvents))
	}
	if mock.ptrEvents[0].x != 400 || mock.ptrEvents[0].y != 300 || mock.ptrEvents[0].buttonMask != 1 {
		t.Errorf("press event = %+v, want {x:400, y:300, buttonMask:1}", mock.ptrEvents[0])
	}
	if mock.ptrEvents[1].x != 400 || mock.ptrEvents[1].y != 300 || mock.ptrEvents[1].buttonMask != 0 {
		t.Errorf("release event = %+v, want {x:400, y:300, buttonMask:0}", mock.ptrEvents[1])
	}
}

func TestSendClick_RightButton(t *testing.T) {
	mock := &mockClient{}
	err := SendClick(mock, 100, 200, 4)
	if err != nil {
		t.Fatalf("SendClick error: %v", err)
	}
	if len(mock.ptrEvents) != 2 {
		t.Fatalf("got %d pointer events, want 2", len(mock.ptrEvents))
	}
	if mock.ptrEvents[0].buttonMask != 4 {
		t.Errorf("press buttonMask = %d, want 4", mock.ptrEvents[0].buttonMask)
	}
	if mock.ptrEvents[1].buttonMask != 0 {
		t.Errorf("release buttonMask = %d, want 0", mock.ptrEvents[1].buttonMask)
	}
}

func TestSendMove(t *testing.T) {
	mock := &mockClient{}
	err := SendMove(mock, 500, 600)
	if err != nil {
		t.Fatalf("SendMove error: %v", err)
	}
	if len(mock.ptrEvents) != 1 {
		t.Fatalf("got %d pointer events, want 1", len(mock.ptrEvents))
	}
	if mock.ptrEvents[0].x != 500 || mock.ptrEvents[0].y != 600 || mock.ptrEvents[0].buttonMask != 0 {
		t.Errorf("move event = %+v, want {x:500, y:600, buttonMask:0}", mock.ptrEvents[0])
	}
}
