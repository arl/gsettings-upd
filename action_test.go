package main

import "testing"

func TestLoadConfig(t *testing.T) {
	var tests = []struct {
		name    string
		cfg     string
		wantErr error
	}{
		{
			name:    "correct config",
			wantErr: nil,
			cfg: `
			{
			  "actions": [
				{
				  "schema": "org.gnome.desktop.screensaver",
				  "key": "picture-uri",
				  "every": "1h",
				  "values": [ "A", "B", "C", "D" ]
				}
			  ]
			}`,
		},
		{
			name:    "action with no values",
			wantErr: errNoValues,
			cfg: `
			{
			  "actions": [
				{
				  "schema": "org.gnome.desktop.screensaver",
				  "key": "picture-uri",
				  "every": "10s",
				  "values": []
				}
			  ]
			}`,
		},
		{
			name:    "empty schema",
			wantErr: errNoSchema,
			cfg: `
			{
			  "actions": [
				{
				  "schema": "", 
				  "key": "picture-uri",
				  "every": "10s",
				  "values": [ "A", "B", "C", "D" ]
				}
			  ]
			}`,
		},
		{
			name:    "empty key",
			wantErr: errNoKey,
			cfg: `
			{
			  "actions": [
				{
				  "schema": "org.gnome.desktop.screensaver",
				  "key": "",
				  "every": "10s",
				  "values": [ "A", "B", "C", "D" ]
				}
			  ]
			}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := loadConfig([]byte(tt.cfg))
			if err != tt.wantErr {
				t.Errorf("newConfig, wantErr %v, got %v", tt.wantErr, err)
			}
		})
	}
}
