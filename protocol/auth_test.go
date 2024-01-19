package protocol

import (
	"net/http"
	"testing"
)

func TestGetCredential(t *testing.T) {
	tests := []struct {
		name          string
		request       *http.Request
		expectedCreds Credential
		expectedErr   error
	}{
		{
			name: "Valid Basic Auth",
			request: &http.Request{
				Header: http.Header{
					"Authorization": []string{"Basic dXNlcm5hbWU6cGFzc3dvcmQ="},
				},
			},
			expectedCreds: Credential{Username: "username", Password: "password"},
			expectedErr:   nil,
		},
		{
			name: "Invalid Basic Auth",
			request: &http.Request{
				Header: http.Header{},
			},
			expectedCreds: Credential{},
			expectedErr:   errAuthenticationFailed,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			creds, err := getCredential(test.request)

			if creds != test.expectedCreds {
				t.Errorf("Expected credentials %v, but got %v", test.expectedCreds, creds)
			}

			if err != test.expectedErr {
				t.Errorf("Expected error %v, but got %v", test.expectedErr, err)
			}
		})
	}
}
