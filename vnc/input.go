package vnc

import "fmt"

func SendKeySequence(client VNCClient, actions []KeyAction) error {
	for _, a := range actions {
		if err := client.SendKey(a.Key, a.Down); err != nil {
			return fmt.Errorf("send key 0x%04x (down=%v): %w", a.Key, a.Down, err)
		}
	}
	return nil
}

func SendTypeString(client VNCClient, text string) error {
	for _, r := range text {
		keysym, shift, err := RuneToKeyInfo(r)
		if err != nil {
			return fmt.Errorf("type string: %w", err)
		}
		if shift {
			if err := client.SendKey(0xffe1, true); err != nil {
				return fmt.Errorf("type string shift press for %q: %w", r, err)
			}
		}
		if err := client.SendKey(keysym, true); err != nil {
			return fmt.Errorf("type string press %q: %w", r, err)
		}
		if err := client.SendKey(keysym, false); err != nil {
			return fmt.Errorf("type string release %q: %w", r, err)
		}
		if shift {
			if err := client.SendKey(0xffe1, false); err != nil {
				return fmt.Errorf("type string shift release for %q: %w", r, err)
			}
		}
	}
	return nil
}

func SendClick(client VNCClient, x, y uint16, buttonMask uint8) error {
	if err := client.SendPointer(x, y, buttonMask); err != nil {
		return fmt.Errorf("click press at (%d,%d): %w", x, y, err)
	}
	if err := client.SendPointer(x, y, 0); err != nil {
		return fmt.Errorf("click release at (%d,%d): %w", x, y, err)
	}
	return nil
}

func SendMove(client VNCClient, x, y uint16) error {
	if err := client.SendPointer(x, y, 0); err != nil {
		return fmt.Errorf("move to (%d,%d): %w", x, y, err)
	}
	return nil
}
