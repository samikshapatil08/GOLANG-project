package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Helper()

	tests := []struct {
		name        string
		envSetup    func()
		expectedPort string
		shouldErr   bool
	}{
		{
			name: "load default values when env is empty",
			envSetup: func() {
				os.Clearenv()
			},
			expectedPort: "8080",
			shouldErr:    false,
		},
		{
			name: "load customized port",
			envSetup: func() {
				os.Clearenv()
				os.Setenv("PORT", "9090")
				os.Setenv("JWT_SECRET", "anothersecretkey")
			},
			expectedPort: "9090",
			shouldErr:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tc.envSetup()
			cfg, err := Load()
			if tc.shouldErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if cfg.Port != tc.expectedPort {
					t.Errorf("expected port %s, got %s", tc.expectedPort, cfg.Port)
				}
			}
		})
	}
}
