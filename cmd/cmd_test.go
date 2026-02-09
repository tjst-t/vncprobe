package cmd

import (
	"testing"
)

func TestParseGlobalFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantAddr string
		wantPass string
		wantTO   int
		wantErr  bool
	}{
		{
			name:     "short server flag",
			args:     []string{"-s", "10.0.0.1:5900"},
			wantAddr: "10.0.0.1:5900",
			wantTO:   10,
		},
		{
			name:     "long server flag",
			args:     []string{"--server", "10.0.0.1:5900"},
			wantAddr: "10.0.0.1:5900",
			wantTO:   10,
		},
		{
			name:     "with password and timeout",
			args:     []string{"-s", "10.0.0.1:5900", "-p", "secret", "--timeout", "30"},
			wantAddr: "10.0.0.1:5900",
			wantPass: "secret",
			wantTO:   30,
		},
		{
			name:    "missing server",
			args:    []string{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts, _, err := ParseGlobalFlags(tt.args)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if opts.Server != tt.wantAddr {
				t.Errorf("Server = %q, want %q", opts.Server, tt.wantAddr)
			}
			if opts.Password != tt.wantPass {
				t.Errorf("Password = %q, want %q", opts.Password, tt.wantPass)
			}
			if opts.Timeout != tt.wantTO {
				t.Errorf("Timeout = %d, want %d", opts.Timeout, tt.wantTO)
			}
		})
	}
}

func TestButtonNumberToMask(t *testing.T) {
	tests := []struct {
		name   string
		button int
		want   uint8
	}{
		{"left (default 1)", 1, 1},
		{"middle (2)", 2, 2},
		{"right (3)", 3, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ButtonNumberToMask(tt.button)
			if got != tt.want {
				t.Errorf("ButtonNumberToMask(%d) = %d, want %d", tt.button, got, tt.want)
			}
		})
	}
}
