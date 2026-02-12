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
	// 'H' = Shift+h (4 events), 'i' = normal (2 events) = 6 total
	want := []KeyAction{
		{Key: 0xffe1, Down: true},  // Shift_L press
		{Key: 0x0068, Down: true},  // 'h' press
		{Key: 0x0068, Down: false}, // 'h' release
		{Key: 0xffe1, Down: false}, // Shift_L release
		{Key: 0x0069, Down: true},  // 'i' press
		{Key: 0x0069, Down: false}, // 'i' release
	}
	if len(mock.keyEvents) != len(want) {
		t.Fatalf("got %d key events, want %d: %+v", len(mock.keyEvents), len(want), mock.keyEvents)
	}
	for i, w := range want {
		if mock.keyEvents[i] != w {
			t.Errorf("event[%d] = {Key:0x%04x, Down:%v}, want {Key:0x%04x, Down:%v}",
				i, mock.keyEvents[i].Key, mock.keyEvents[i].Down, w.Key, w.Down)
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

func TestSendTypeString_ShiftedChars(t *testing.T) {
	mock := &mockClient{}
	err := SendTypeString(mock, "!")
	if err != nil {
		t.Fatalf("SendTypeString error: %v", err)
	}
	// '!' requires Shift+1, so expect: Shift press, '1' press, '1' release, Shift release
	want := []KeyAction{
		{Key: 0xffe1, Down: true},  // Shift_L press
		{Key: 0x0031, Down: true},  // '1' press
		{Key: 0x0031, Down: false}, // '1' release
		{Key: 0xffe1, Down: false}, // Shift_L release
	}
	if len(mock.keyEvents) != len(want) {
		t.Fatalf("got %d key events, want %d: %+v", len(mock.keyEvents), len(want), mock.keyEvents)
	}
	for i, w := range want {
		if mock.keyEvents[i] != w {
			t.Errorf("event[%d] = %+v, want %+v", i, mock.keyEvents[i], w)
		}
	}
}

func TestSendTypeString_AllShiftedSymbols(t *testing.T) {
	// Test all shifted characters from US keyboard layout
	tests := []struct {
		char    string
		baseKey uint32 // the unshifted key
	}{
		{"!", 0x0031}, // Shift+1
		{"@", 0x0032}, // Shift+2
		{"#", 0x0033}, // Shift+3
		{"$", 0x0034}, // Shift+4
		{"%", 0x0035}, // Shift+5
		{"^", 0x0036}, // Shift+6
		{"&", 0x0037}, // Shift+7
		{"*", 0x0038}, // Shift+8
		{"(", 0x0039}, // Shift+9
		{")", 0x0030}, // Shift+0
		{"_", 0x002d}, // Shift+-
		{"+", 0x003d}, // Shift+=
		{"{", 0x005b}, // Shift+[
		{"}", 0x005d}, // Shift+]
		{"|", 0x005c}, // Shift+\
		{":", 0x003b}, // Shift+;
		{"\"", 0x0027}, // Shift+'
		{"<", 0x002c}, // Shift+,
		{">", 0x002e}, // Shift+.
		{"?", 0x002f}, // Shift+/
		{"~", 0x0060}, // Shift+`
	}
	for _, tt := range tests {
		t.Run(tt.char, func(t *testing.T) {
			mock := &mockClient{}
			err := SendTypeString(mock, tt.char)
			if err != nil {
				t.Fatalf("SendTypeString(%q) error: %v", tt.char, err)
			}
			want := []KeyAction{
				{Key: 0xffe1, Down: true},   // Shift_L press
				{Key: tt.baseKey, Down: true},  // base key press
				{Key: tt.baseKey, Down: false}, // base key release
				{Key: 0xffe1, Down: false},  // Shift_L release
			}
			if len(mock.keyEvents) != len(want) {
				t.Fatalf("got %d events, want %d: %+v", len(mock.keyEvents), len(want), mock.keyEvents)
			}
			for i, w := range want {
				if mock.keyEvents[i] != w {
					t.Errorf("event[%d] = {Key:0x%04x, Down:%v}, want {Key:0x%04x, Down:%v}",
						i, mock.keyEvents[i].Key, mock.keyEvents[i].Down, w.Key, w.Down)
				}
			}
		})
	}
}

func TestSendTypeString_UppercaseLetters(t *testing.T) {
	mock := &mockClient{}
	err := SendTypeString(mock, "A")
	if err != nil {
		t.Fatalf("SendTypeString error: %v", err)
	}
	// 'A' requires Shift+a
	want := []KeyAction{
		{Key: 0xffe1, Down: true},  // Shift_L press
		{Key: 0x0061, Down: true},  // 'a' press
		{Key: 0x0061, Down: false}, // 'a' release
		{Key: 0xffe1, Down: false}, // Shift_L release
	}
	if len(mock.keyEvents) != len(want) {
		t.Fatalf("got %d key events, want %d: %+v", len(mock.keyEvents), len(want), mock.keyEvents)
	}
	for i, w := range want {
		if mock.keyEvents[i] != w {
			t.Errorf("event[%d] = %+v, want %+v", i, mock.keyEvents[i], w)
		}
	}
}

func TestSendTypeString_MixedShiftAndNormal(t *testing.T) {
	mock := &mockClient{}
	err := SendTypeString(mock, "a!")
	if err != nil {
		t.Fatalf("SendTypeString error: %v", err)
	}
	// 'a' = normal, '!' = Shift+1
	want := []KeyAction{
		{Key: 0x0061, Down: true},  // 'a' press
		{Key: 0x0061, Down: false}, // 'a' release
		{Key: 0xffe1, Down: true},  // Shift_L press
		{Key: 0x0031, Down: true},  // '1' press
		{Key: 0x0031, Down: false}, // '1' release
		{Key: 0xffe1, Down: false}, // Shift_L release
	}
	if len(mock.keyEvents) != len(want) {
		t.Fatalf("got %d key events, want %d: %+v", len(mock.keyEvents), len(want), mock.keyEvents)
	}
	for i, w := range want {
		if mock.keyEvents[i] != w {
			t.Errorf("event[%d] = {Key:0x%04x, Down:%v}, want {Key:0x%04x, Down:%v}",
				i, mock.keyEvents[i].Key, mock.keyEvents[i].Down, w.Key, w.Down)
		}
	}
}
