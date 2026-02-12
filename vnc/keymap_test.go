package vnc

import (
	"testing"
)

func TestParseKeySequence_SingleKeys(t *testing.T) {
	tests := []struct {
		input string
		want  []KeyAction
	}{
		{"enter", []KeyAction{{Key: 0xff0d, Down: true}, {Key: 0xff0d, Down: false}}},
		{"tab", []KeyAction{{Key: 0xff09, Down: true}, {Key: 0xff09, Down: false}}},
		{"escape", []KeyAction{{Key: 0xff1b, Down: true}, {Key: 0xff1b, Down: false}}},
		{"backspace", []KeyAction{{Key: 0xff08, Down: true}, {Key: 0xff08, Down: false}}},
		{"delete", []KeyAction{{Key: 0xffff, Down: true}, {Key: 0xffff, Down: false}}},
		{"space", []KeyAction{{Key: 0x0020, Down: true}, {Key: 0x0020, Down: false}}},
		{"up", []KeyAction{{Key: 0xff52, Down: true}, {Key: 0xff52, Down: false}}},
		{"down", []KeyAction{{Key: 0xff54, Down: true}, {Key: 0xff54, Down: false}}},
		{"left", []KeyAction{{Key: 0xff51, Down: true}, {Key: 0xff51, Down: false}}},
		{"right", []KeyAction{{Key: 0xff53, Down: true}, {Key: 0xff53, Down: false}}},
		{"home", []KeyAction{{Key: 0xff50, Down: true}, {Key: 0xff50, Down: false}}},
		{"end", []KeyAction{{Key: 0xff57, Down: true}, {Key: 0xff57, Down: false}}},
		{"pageup", []KeyAction{{Key: 0xff55, Down: true}, {Key: 0xff55, Down: false}}},
		{"pagedown", []KeyAction{{Key: 0xff56, Down: true}, {Key: 0xff56, Down: false}}},
		{"f1", []KeyAction{{Key: 0xffbe, Down: true}, {Key: 0xffbe, Down: false}}},
		{"f12", []KeyAction{{Key: 0xffc9, Down: true}, {Key: 0xffc9, Down: false}}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseKeySequence(tt.input)
			if err != nil {
				t.Fatalf("ParseKeySequence(%q) error: %v", tt.input, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseKeySequence(%q) got %d actions, want %d", tt.input, len(got), len(tt.want))
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("ParseKeySequence(%q)[%d] = %+v, want %+v", tt.input, i, g, tt.want[i])
				}
			}
		})
	}
}

func TestParseKeySequence_ModifierCombos(t *testing.T) {
	tests := []struct {
		input string
		want  []KeyAction
	}{
		{"ctrl-c", []KeyAction{
			{Key: 0xffe3, Down: true},
			{Key: 0x0063, Down: true},
			{Key: 0x0063, Down: false},
			{Key: 0xffe3, Down: false},
		}},
		{"alt-f4", []KeyAction{
			{Key: 0xffe9, Down: true},
			{Key: 0xffc1, Down: true},
			{Key: 0xffc1, Down: false},
			{Key: 0xffe9, Down: false},
		}},
		{"shift-a", []KeyAction{
			{Key: 0xffe1, Down: true},
			{Key: 0x0061, Down: true},
			{Key: 0x0061, Down: false},
			{Key: 0xffe1, Down: false},
		}},
		{"ctrl-alt-delete", []KeyAction{
			{Key: 0xffe3, Down: true},
			{Key: 0xffe9, Down: true},
			{Key: 0xffff, Down: true},
			{Key: 0xffff, Down: false},
			{Key: 0xffe9, Down: false},
			{Key: 0xffe3, Down: false},
		}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseKeySequence(tt.input)
			if err != nil {
				t.Fatalf("ParseKeySequence(%q) error: %v", tt.input, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("ParseKeySequence(%q) got %d actions, want %d\ngot:  %+v\nwant: %+v", tt.input, len(got), len(tt.want), got, tt.want)
			}
			for i, g := range got {
				if g != tt.want[i] {
					t.Errorf("ParseKeySequence(%q)[%d] = %+v, want %+v", tt.input, i, g, tt.want[i])
				}
			}
		})
	}
}

func TestParseKeySequence_SingleChar(t *testing.T) {
	got, err := ParseKeySequence("a")
	if err != nil {
		t.Fatalf("ParseKeySequence(%q) error: %v", "a", err)
	}
	want := []KeyAction{
		{Key: 0x0061, Down: true},
		{Key: 0x0061, Down: false},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d actions, want %d", len(got), len(want))
	}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("[%d] = %+v, want %+v", i, g, want[i])
		}
	}
}

func TestParseKeySequence_UnknownKey(t *testing.T) {
	_, err := ParseKeySequence("nonexistentkey")
	if err == nil {
		t.Fatal("ParseKeySequence(\"nonexistentkey\") expected error, got nil")
	}
}

func TestParseKeySequence_ShiftedChars(t *testing.T) {
	// When a single shifted character is passed to ParseKeySequence,
	// it should produce Shift_L press + base key press/release + Shift_L release
	tests := []struct {
		input   string
		wantKey uint32 // expected base key
	}{
		{"!", 0x0031},  // Shift+1
		{"@", 0x0032},  // Shift+2
		{"#", 0x0033},  // Shift+3
		{"$", 0x0034},  // Shift+4
		{"%", 0x0035},  // Shift+5
		{"^", 0x0036},  // Shift+6
		{"&", 0x0037},  // Shift+7
		{"*", 0x0038},  // Shift+8
		{"(", 0x0039},  // Shift+9
		{")", 0x0030},  // Shift+0
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseKeySequence(tt.input)
			if err != nil {
				t.Fatalf("ParseKeySequence(%q) error: %v", tt.input, err)
			}
			want := []KeyAction{
				{Key: 0xffe1, Down: true},     // Shift_L press
				{Key: tt.wantKey, Down: true},  // base key press
				{Key: tt.wantKey, Down: false}, // base key release
				{Key: 0xffe1, Down: false},    // Shift_L release
			}
			if len(got) != len(want) {
				t.Fatalf("ParseKeySequence(%q) got %d actions, want %d\ngot:  %+v\nwant: %+v",
					tt.input, len(got), len(want), got, want)
			}
			for i, g := range got {
				if g != want[i] {
					t.Errorf("ParseKeySequence(%q)[%d] = %+v, want %+v", tt.input, i, g, want[i])
				}
			}
		})
	}
}

func TestParseKeySequence_UppercaseLetter(t *testing.T) {
	// Uppercase 'A' should produce Shift+a
	// Note: ParseKeySequence lowercases input, but we should still handle
	// the case where the original character is uppercase
	got, err := ParseKeySequence("A")
	if err != nil {
		t.Fatalf("ParseKeySequence(\"A\") error: %v", err)
	}
	want := []KeyAction{
		{Key: 0xffe1, Down: true},  // Shift_L press
		{Key: 0x0061, Down: true},  // 'a' press
		{Key: 0x0061, Down: false}, // 'a' release
		{Key: 0xffe1, Down: false}, // Shift_L release
	}
	if len(got) != len(want) {
		t.Fatalf("got %d actions, want %d\ngot:  %+v\nwant: %+v", len(got), len(want), got, want)
	}
	for i, g := range got {
		if g != want[i] {
			t.Errorf("[%d] = %+v, want %+v", i, g, want[i])
		}
	}
}

func TestRuneToKeyCode(t *testing.T) {
	// RuneToKeyCode returns the base keysym (without shift info)
	tests := []struct {
		r    rune
		want uint32
	}{
		{'a', 0x0061},
		{'A', 0x0061}, // base key for 'A' is 'a'
		{'0', 0x0030},
		{' ', 0x0020},
		{'!', 0x0031}, // base key for '!' is '1'
	}
	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			got, err := RuneToKeyCode(tt.r)
			if err != nil {
				t.Fatalf("RuneToKeyCode(%q) error: %v", tt.r, err)
			}
			if got != tt.want {
				t.Errorf("RuneToKeyCode(%q) = 0x%04x, want 0x%04x", tt.r, got, tt.want)
			}
		})
	}
}

func TestRuneToKeyInfo(t *testing.T) {
	tests := []struct {
		r         rune
		wantKey   uint32
		wantShift bool
	}{
		{'a', 0x0061, false},
		{'A', 0x0061, true},
		{'z', 0x007a, false},
		{'Z', 0x007a, true},
		{'0', 0x0030, false},
		{' ', 0x0020, false},
		{'!', 0x0031, true},
		{'@', 0x0032, true},
		{'-', 0x002d, false},
		{'_', 0x002d, true},
		{'=', 0x003d, false},
		{'+', 0x003d, true},
	}
	for _, tt := range tests {
		t.Run(string(tt.r), func(t *testing.T) {
			key, shift, err := RuneToKeyInfo(tt.r)
			if err != nil {
				t.Fatalf("RuneToKeyInfo(%q) error: %v", tt.r, err)
			}
			if key != tt.wantKey {
				t.Errorf("RuneToKeyInfo(%q) key = 0x%04x, want 0x%04x", tt.r, key, tt.wantKey)
			}
			if shift != tt.wantShift {
				t.Errorf("RuneToKeyInfo(%q) shift = %v, want %v", tt.r, shift, tt.wantShift)
			}
		})
	}
}
