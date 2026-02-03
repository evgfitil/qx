package guard

import (
	"testing"
)

func TestSanitizer_Check(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantSecret bool
		wantRuleID string
	}{
		// AWS Access Key
		{
			name:       "aws access key AKIA",
			input:      "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE",
			wantSecret: true,
			wantRuleID: "aws-access-key",
		},
		{
			name:       "aws access key ASIA",
			input:      "ASIAISOMETHINGFODN7X",
			wantSecret: true,
			wantRuleID: "aws-access-key",
		},
		{
			name:       "not aws key - wrong prefix",
			input:      "AKXXIOSFODNN7EXAMPLE",
			wantSecret: false,
		},

		// AWS Secret Key
		{
			name:       "aws secret key",
			input:      `aws_secret_access_key="wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"`,
			wantSecret: true,
			wantRuleID: "aws-secret-key",
		},

		// OpenAI API Key
		{
			name:       "openai api key",
			input:      "sk-" + "x]x[x]x[x]x[x]x[x]x[" + "T3BlbkFJ" + "x]x[x]x[x]x[x]x[x]x[",
			wantSecret: false,
		},
		{
			name:       "openai project key",
			input:      "sk-proj-" + string(make([]byte, 74)) + "T3BlbkFJ" + string(make([]byte, 20)),
			wantSecret: false, // bytes are null, won't match pattern
		},

		// GitHub PAT
		{
			name:       "github pat",
			input:      "ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			wantSecret: true,
			wantRuleID: "github-pat",
		},
		{
			name:       "github pat in curl",
			input:      `curl -H "Authorization: Bearer ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"`,
			wantSecret: true,
			wantRuleID: "github-pat",
		},
		{
			name:       "not github pat - too short",
			input:      "ghp_xxxxxxxxxxxxxxxxxx",
			wantSecret: false,
		},

		// GitHub App Token
		{
			name:       "github app token ghu",
			input:      "ghu_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			wantSecret: true,
			wantRuleID: "github-app-token",
		},
		{
			name:       "github app token ghs",
			input:      "ghs_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
			wantSecret: true,
			wantRuleID: "github-app-token",
		},

		// GitLab PAT
		{
			name:       "gitlab pat",
			input:      "glpat-xxxxxxxxxxxxxxxxxxxx",
			wantSecret: true,
			wantRuleID: "gitlab-pat",
		},

		// Private Key
		{
			name:       "rsa private key",
			input:      "-----BEGIN RSA PRIVATE KEY-----",
			wantSecret: true,
			wantRuleID: "private-key",
		},
		{
			name:       "openssh private key",
			input:      "-----BEGIN OPENSSH PRIVATE KEY-----",
			wantSecret: true,
			wantRuleID: "private-key",
		},

		// JWT
		{
			name:       "jwt token",
			input:      "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
			wantSecret: true,
			wantRuleID: "jwt",
		},

		// Generic API Key
		{
			name:       "generic api_key",
			input:      `api_key="someVeryLongApiKeyValue1234"`,
			wantSecret: true,
			wantRuleID: "generic-api-key",
		},
		{
			name:       "generic access_token",
			input:      "access_token:abcdefghijklmnopqrstuvwxyz",
			wantSecret: true,
			wantRuleID: "generic-api-key",
		},

		// Password in URL
		{
			name:       "mongodb url with password",
			input:      "mongodb://admin:secretpassword@localhost:27017",
			wantSecret: true,
			wantRuleID: "password-in-url",
		},
		{
			name:       "postgresql url with password",
			input:      "postgresql://user:pass123@db.example.com:5432/mydb",
			wantSecret: true,
			wantRuleID: "password-in-url",
		},
		{
			name:       "mysql url with password",
			input:      "mysql://root:password@localhost/database",
			wantSecret: true,
			wantRuleID: "password-in-url",
		},
		{
			name:       "redis url with password",
			input:      "redis://default:mypassword@cache.example.com:6379",
			wantSecret: true,
			wantRuleID: "password-in-url",
		},

		// Bearer Token
		{
			name:       "bearer token",
			input:      "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6",
			wantSecret: true,
			wantRuleID: "bearer-token",
		},

		// Edge cases
		{
			name:       "empty string",
			input:      "",
			wantSecret: false,
		},
		{
			name:       "regular command",
			input:      "ls -la /home/user",
			wantSecret: false,
		},
		{
			name:       "unicode text",
			input:      "echo '日本語テキスト'",
			wantSecret: false,
		},
		{
			name:       "long query without secrets",
			input:      "find /var/log -name '*.log' -mtime +30 -exec rm {} \\; && echo 'cleaned up old logs'",
			wantSecret: false,
		},
	}

	s := New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := s.Check(tt.input)
			if result.HasSecrets != tt.wantSecret {
				t.Errorf("Check() HasSecrets = %v, want %v", result.HasSecrets, tt.wantSecret)
			}
			if tt.wantSecret && tt.wantRuleID != "" {
				found := false
				for _, d := range result.Detected {
					if d.RuleID == tt.wantRuleID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Check() did not detect rule %s, got %v", tt.wantRuleID, result.Detected)
				}
			}
		})
	}
}

func TestSanitizeOutput(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal command",
			input:    "ls -la",
			expected: "ls -la",
		},
		{
			name:     "escape sequence stripped",
			input:    "ls\x1b[2Drm",
			expected: "ls[2Drm",
		},
		{
			name:     "null byte stripped",
			input:    "echo\x00hidden",
			expected: "echohidden",
		},
		{
			name:     "DEL and control chars stripped",
			input:    "cmd\x7f\x1f",
			expected: "cmd",
		},
		{
			name:     "preserve tab",
			input:    "echo\t'hello'",
			expected: "echo\t'hello'",
		},
		{
			name:     "preserve newline",
			input:    "echo 'line1'\necho 'line2'",
			expected: "echo 'line1'\necho 'line2'",
		},
		{
			name:     "multiline with tabs",
			input:    "if true; then\n\techo 'yes'\nfi",
			expected: "if true; then\n\techo 'yes'\nfi",
		},
		{
			name:     "bell character stripped",
			input:    "echo\x07alert",
			expected: "echoalert",
		},
		{
			name:     "backspace stripped",
			input:    "rm -rf\x08\x08ignore",
			expected: "rm -rfignore",
		},
		{
			name:     "form feed stripped",
			input:    "page1\x0cpage2",
			expected: "page1page2",
		},
		{
			name:     "vertical tab stripped",
			input:    "col1\x0bcol2",
			expected: "col1col2",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "unicode preserved",
			input:    "echo '日本語'",
			expected: "echo '日本語'",
		},
		{
			name:     "ansi color codes stripped",
			input:    "\x1b[31mred\x1b[0m",
			expected: "[31mred[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SanitizeOutput(tt.input)
			if got != tt.expected {
				t.Errorf("SanitizeOutput(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSanitizer_MultipleSecrets(t *testing.T) {
	s := New()
	input := "export AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE && curl -H 'Authorization: Bearer ghp_xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx'"

	result := s.Check(input)
	if !result.HasSecrets {
		t.Error("expected HasSecrets to be true")
	}
	if len(result.Detected) < 2 {
		t.Errorf("expected at least 2 detections, got %d", len(result.Detected))
	}
}

func TestIsExplanation(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid command - ls",
			input:    "ls -la",
			expected: false,
		},
		{
			name:     "valid command - docker",
			input:    "docker ps -a",
			expected: false,
		},
		{
			name:     "valid command - find",
			input:    "find . -name '*.go' -exec grep TODO {} +",
			expected: false,
		},
		{
			name:     "explanation - the command",
			input:    "The command ls -la lists all files",
			expected: true,
		},
		{
			name:     "explanation - this command",
			input:    "This command will show you the files",
			expected: true,
		},
		{
			name:     "explanation - this will",
			input:    "This will list all files in the directory",
			expected: true,
		},
		{
			name:     "explanation - here is",
			input:    "Here is the command you need: ls -la",
			expected: true,
		},
		{
			name:     "explanation - here's",
			input:    "Here's how to do it: ls -la",
			expected: true,
		},
		{
			name:     "explanation - you can use",
			input:    "You can use the following command",
			expected: true,
		},
		{
			name:     "explanation - i suggest",
			input:    "I suggest using docker ps instead",
			expected: true,
		},
		{
			name:     "explanation - note",
			input:    "Note: this command requires sudo",
			expected: true,
		},
		{
			name:     "explanation - the following",
			input:    "The following command will help",
			expected: true,
		},
		{
			name:     "explanation - i recommend",
			input:    "I recommend running this first",
			expected: true,
		},
		{
			name:     "explanation - lowercase the command",
			input:    "the command lists files",
			expected: true,
		},
		{
			name:     "explanation - with leading whitespace",
			input:    "  The command ls lists files",
			expected: true,
		},
		{
			name:     "valid command - echo the command",
			input:    "echo 'the command'",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: false,
		},
		{
			name:     "valid command - pipe",
			input:    "docker ps | grep nginx",
			expected: false,
		},
		{
			name:     "valid command - multiline",
			input:    "ls -la && echo 'done'",
			expected: false,
		},
		{
			name:     "explanation - this would",
			input:    "This would display all running containers",
			expected: true,
		},
		{
			name:     "explanation - to do this",
			input:    "To do this, run the following command",
			expected: true,
		},
		{
			name:     "explanation - you could use",
			input:    "You could use grep for searching",
			expected: true,
		},
		{
			name:     "explanation - you should",
			input:    "You should run this as root",
			expected: true,
		},
		{
			name:     "explanation - i would",
			input:    "I would recommend using docker",
			expected: true,
		},
		{
			name:     "explanation - i'd suggest",
			input:    "I'd suggest trying this approach",
			expected: true,
		},
		{
			name:     "explanation - explanation prefix",
			input:    "Explanation: the command does X",
			expected: true,
		},
		{
			name:     "explanation - below is",
			input:    "Below is the command you need",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExplanation(tt.input)
			if got != tt.expected {
				t.Errorf("IsExplanation(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}
