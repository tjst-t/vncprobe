package vnc

import (
	"fmt"
	"strings"
)

type KeyAction struct {
	Key  uint32
	Down bool
}

var namedKeys = map[string]uint32{
	"enter":     0xff0d,
	"return":    0xff0d,
	"tab":       0xff09,
	"escape":    0xff1b,
	"esc":       0xff1b,
	"backspace": 0xff08,
	"delete":    0xffff,
	"space":     0x0020,
	"up":        0xff52,
	"down":      0xff54,
	"left":      0xff51,
	"right":     0xff53,
	"home":      0xff50,
	"end":       0xff57,
	"pageup":    0xff55,
	"pagedown":  0xff56,
	"insert":    0xff63,
	"f1":        0xffbe,
	"f2":        0xffbf,
	"f3":        0xffc0,
	"f4":        0xffc1,
	"f5":        0xffc2,
	"f6":        0xffc3,
	"f7":        0xffc4,
	"f8":        0xffc5,
	"f9":        0xffc6,
	"f10":       0xffc7,
	"f11":       0xffc8,
	"f12":       0xffc9,
}

var modifierKeys = map[string]uint32{
	"ctrl":  0xffe3,
	"alt":   0xffe9,
	"shift": 0xffe1,
	"super": 0xffeb,
	"meta":  0xffe7,
}

// shiftedCharToBase maps shifted characters to their unshifted base key (US keyboard layout).
var shiftedCharToBase = map[rune]rune{
	'!': '1', '@': '2', '#': '3', '$': '4', '%': '5',
	'^': '6', '&': '7', '*': '8', '(': '9', ')': '0',
	'_': '-', '+': '=',
	'{': '[', '}': ']', '|': '\\',
	':': ';', '"': '\'',
	'<': ',', '>': '.', '?': '/',
	'~': '`',
}

// RuneToKeyInfo returns the base keysym and whether Shift is needed for the given rune.
func RuneToKeyInfo(r rune) (keysym uint32, shift bool, err error) {
	// Uppercase letters: Shift + lowercase
	if r >= 'A' && r <= 'Z' {
		return uint32(r - 'A' + 'a'), true, nil
	}
	// Shifted symbols: Shift + base key
	if base, ok := shiftedCharToBase[r]; ok {
		return uint32(base), true, nil
	}
	// Normal printable ASCII
	if r >= 0x20 && r <= 0x7e {
		return uint32(r), false, nil
	}
	return 0, false, fmt.Errorf("unsupported rune: %q (0x%04x)", r, r)
}

func RuneToKeyCode(r rune) (uint32, error) {
	keysym, _, err := RuneToKeyInfo(r)
	return keysym, err
}

func ParseKeySequence(input string) ([]KeyAction, error) {
	lower := strings.ToLower(input)
	parts := strings.Split(lower, "-")

	var modifiers []uint32
	finalKeyStr := ""
	finalKeyOriginal := ""

	for i, part := range parts {
		if code, ok := modifierKeys[part]; ok {
			modifiers = append(modifiers, code)
		} else {
			finalKeyStr = strings.Join(parts[i:], "-")
			// Reconstruct original-case version for single-char handling
			origParts := strings.SplitN(input, "-", i+1)
			if len(origParts) > i {
				finalKeyOriginal = origParts[i]
			} else {
				finalKeyOriginal = finalKeyStr
			}
			break
		}
	}

	if finalKeyStr == "" {
		if len(modifiers) == 0 {
			return nil, fmt.Errorf("empty key sequence")
		}
		lastMod := modifiers[len(modifiers)-1]
		modifiers = modifiers[:len(modifiers)-1]
		var actions []KeyAction
		for _, m := range modifiers {
			actions = append(actions, KeyAction{Key: m, Down: true})
		}
		actions = append(actions, KeyAction{Key: lastMod, Down: true})
		actions = append(actions, KeyAction{Key: lastMod, Down: false})
		for i := len(modifiers) - 1; i >= 0; i-- {
			actions = append(actions, KeyAction{Key: modifiers[i], Down: false})
		}
		return actions, nil
	}

	var finalKeyCode uint32
	if code, ok := namedKeys[finalKeyStr]; ok {
		finalKeyCode = code
	} else if len(finalKeyStr) == 1 {
		r := rune(finalKeyOriginal[0])
		keysym, shift, err := RuneToKeyInfo(r)
		if err != nil {
			return nil, err
		}
		finalKeyCode = keysym
		if shift {
			modifiers = append([]uint32{0xffe1}, modifiers...)
		}
	} else {
		return nil, fmt.Errorf("unknown key: %q", finalKeyStr)
	}

	var actions []KeyAction
	for _, m := range modifiers {
		actions = append(actions, KeyAction{Key: m, Down: true})
	}
	actions = append(actions, KeyAction{Key: finalKeyCode, Down: true})
	actions = append(actions, KeyAction{Key: finalKeyCode, Down: false})
	for i := len(modifiers) - 1; i >= 0; i-- {
		actions = append(actions, KeyAction{Key: modifiers[i], Down: false})
	}
	return actions, nil
}
