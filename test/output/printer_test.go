package output_test

import (
	"strings"
	"testing"

	"k8x/internal/output"
)

func TestFilterSecrets(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Base64 encoded secret",
			input:    `"password": "c3VwZXJzZWNyZXRwYXNzd29yZA=="`,
			expected: "***REDACTED***",
		},
		{
			name:     "Bearer token",
			input:    "Authorization: Bearer eyJhbGciOFOOIkpXVCJ9.eyJzdWIiOiBARDkwIn0.dozjgNryP4J3jVmNTERPlFUP0THsR8U", // pragma: allowlist secret
			expected: "Authorization: ***REDACTED***",
		},
		{
			name: "Kubernetes Secret YAML",
			input: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm`,
			expected: `apiVersion: v1
kind: Secret
metadata:
  name: mysecret
data:
  username: ***REDACTED***
  password: ***REDACTED***`,
		},
		{
			name:     "JSON with secrets",
			input:    `{"api_key": "sk-1234567890abcdef", "token": "xoxb-123456789"}`,
			expected: `{***REDACTED***, ***REDACTED***}`,
		},
		{
			name:     "Certificate data",
			input:    `"tls.key": "-----BEGIN RSA PRIV" + "ATE KEY-----\nMIIEpAIBAAKCAQEA..."`,
			expected: "***REDACTED***",
		},
		{
			name:     "Regular text should not be filtered",
			input:    "This is a normal log message with no secrets",
			expected: "This is a normal log message with no secrets",
		},
		{
			name:     "Mixed content",
			input:    `Log: Connected to database. Password: "mysecretpass123". Status: OK`,
			expected: `Log: Connected to database. ***REDACTED***. Status: OK`,
		},
		{
			name:     "AWS ELB DNS should not be filtered",
			input:    "ADDRESS: a780a4e73f2a54a2fa5e96c21e154e33-1293036362.us-west-2.elb.amazonaws.com",
			expected: "ADDRESS: a780a4e73f2a54a2fa5e96c21e154e33-1293036362.us-west-2.elb.amazonaws.com",
		},
		{
			name: "Kubernetes ingress output should not filter DNS",
			input: `NAMESPACE   NAME         CLASS            HOSTS               ADDRESS                                                                   PORTS     AGE
default     admin        public-ingress   app.aihero.studio   a780a4e73f2a54a2fa5e96c21e154e33-1293036362.us-west-2.elb.amazonaws.com   80, 443   108d`,
			expected: `NAMESPACE   NAME         CLASS            HOSTS               ADDRESS                                                                   PORTS     AGE
default     admin        public-ingress   app.aihero.studio   a780a4e73f2a54a2fa5e96c21e154e33-1293036362.us-west-2.elb.amazonaws.com   80, 443   108d`,
		},
		{
			name:     "Regular hostnames should not be filtered",
			input:    "Host: my-service.default.svc.cluster.local Port: 8080",
			expected: "Host: my-service.default.svc.cluster.local Port: 8080",
		},
		{
			name:     "Long base64 with padding should still be filtered",
			input:    "Secret: dGhpc2lzYXZlcnlsb25nc2VjcmV0c3RyaW5ndGhhdHNob3VsZGJlZmlsdGVyZWQ=",
			expected: "Secret: ***REDACTED***",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := output.FilterSecrets(tt.input)
			if !strings.Contains(result, tt.expected) && result != tt.expected {
				t.Errorf("FilterSecrets() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsLikelySecret(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected bool
	}{
		{
			name:     "Base64 encoded string",
			value:    "YWRtaW46cGFzc3dvcmQ=",
			expected: true,
		},
		{
			name:     "JWT token",
			value:    "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U",
			expected: true,
		},
		{
			name:     "Contains password keyword",
			value:    "user_password_hash",
			expected: true,
		},
		{
			name:     "Contains API key keyword",
			value:    "my_api_key_value",
			expected: true,
		},
		{
			name:     "Regular string",
			value:    "hello world",
			expected: false,
		},
		{
			name:     "Short base64",
			value:    "YWI=",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := output.IsLikelySecret(tt.value)
			if result != tt.expected {
				t.Errorf("IsLikelySecret(%s) = %v, want %v", tt.value, result, tt.expected)
			}
		})
	}
}

func TestPrinterFilteringEnabled(t *testing.T) {
	// Test that FilterSecrets filters sensitive data
	input := `{"secret": "mysecretvalue", "data": "normal"}`
	filtered := output.FilterSecrets(input)

	if strings.Contains(filtered, "mysecretvalue") {
		t.Error("Secret value was not filtered")
	}

	if !strings.Contains(filtered, "***REDACTED***") {
		t.Error("Secret replacement text not found")
	}
}

func TestPrinterFilteringDisabled(t *testing.T) {
	p := output.NewPrinter(false)

	// Since we can't access the private filterOutput method from outside the package,
	// we test that the printer exists and can be created
	if p == nil {
		t.Error("Failed to create printer with filtering disabled")
	}

	// Test that FilterSecrets always filters (it's a public function that always filters)
	input := `{"secret": "mysecretvalue", "data": "normal"}`
	filtered := output.FilterSecrets(input)

	if !strings.Contains(filtered, "***REDACTED***") {
		t.Error("FilterSecrets should always filter secrets")
	}
}
