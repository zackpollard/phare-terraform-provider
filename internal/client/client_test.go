// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"testing"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		apiToken  string
		baseURL   string
		wantError bool
	}{
		{
			name:      "valid client with token and URL",
			apiToken:  "test-token",
			baseURL:   "https://api.example.com",
			wantError: false,
		},
		{
			name:      "valid client with token and default URL",
			apiToken:  "test-token",
			baseURL:   "",
			wantError: false,
		},
		{
			name:      "invalid client without token",
			apiToken:  "",
			baseURL:   "https://api.example.com",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.apiToken, tt.baseURL)

			if tt.wantError {
				if err == nil {
					t.Errorf("NewClient() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("NewClient() unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("NewClient() returned nil client")
				return
			}

			if client.apiToken != tt.apiToken {
				t.Errorf("NewClient() apiToken = %v, want %v", client.apiToken, tt.apiToken)
			}

			expectedURL := tt.baseURL
			if expectedURL == "" {
				expectedURL = DefaultBaseURL
			}
			if client.baseURL != expectedURL {
				t.Errorf("NewClient() baseURL = %v, want %v", client.baseURL, expectedURL)
			}
		})
	}
}
