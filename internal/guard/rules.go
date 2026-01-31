package guard

import "regexp"

// Rule defines a secret detection pattern (naming follows gitleaks convention).
type Rule struct {
	RuleID      string
	Description string
	Regex       *regexp.Regexp
}

// DefaultRules contains patterns for common secrets based on gitleaks.
var DefaultRules = []Rule{
	{
		RuleID:      "aws-access-key",
		Description: "AWS Access Key",
		Regex:       regexp.MustCompile(`\b((?:A3T[A-Z0-9]|AKIA|ASIA|ABIA|ACCA)[A-Z2-7]{16})\b`),
	},
	{
		RuleID:      "aws-secret-key",
		Description: "AWS Secret Key",
		Regex:       regexp.MustCompile(`(?i)(aws_secret_access_key|aws_secret_key)[=:]["']?([A-Za-z0-9/+=]{40})["']?`),
	},
	{
		RuleID:      "openai-api-key",
		Description: "OpenAI API Key",
		Regex:       regexp.MustCompile(`\b(sk-(?:proj|svcacct|admin)-[A-Za-z0-9_-]{74,}T3BlbkFJ[A-Za-z0-9_-]{20,}|sk-[a-zA-Z0-9]{20}T3BlbkFJ[a-zA-Z0-9]{20})\b`),
	},
	{
		RuleID:      "anthropic-api-key",
		Description: "Anthropic API Key",
		Regex:       regexp.MustCompile(`\b(sk-ant-api03-[a-zA-Z0-9_\-]{93}AA)\b`),
	},
	{
		RuleID:      "github-pat",
		Description: "GitHub Personal Access Token",
		Regex:       regexp.MustCompile(`ghp_[0-9a-zA-Z]{36}`),
	},
	{
		RuleID:      "github-fine-grained-pat",
		Description: "GitHub Fine-Grained PAT",
		Regex:       regexp.MustCompile(`github_pat_\w{82}`),
	},
	{
		RuleID:      "github-app-token",
		Description: "GitHub App Token",
		Regex:       regexp.MustCompile(`(?:ghu|ghs)_[0-9a-zA-Z]{36}`),
	},
	{
		RuleID:      "gitlab-pat",
		Description: "GitLab Personal Access Token",
		Regex:       regexp.MustCompile(`glpat-[0-9a-zA-Z\-_]{20}`),
	},
	{
		RuleID:      "private-key",
		Description: "Private Key",
		Regex:       regexp.MustCompile(`(?i)-----BEGIN[ A-Z0-9_-]{0,100}PRIVATE KEY-----`),
	},
	{
		RuleID:      "jwt",
		Description: "JWT Token",
		Regex:       regexp.MustCompile(`\b(ey[a-zA-Z0-9]{17,}\.ey[a-zA-Z0-9\/\\_-]{17,}\.[a-zA-Z0-9\/\\_-]{10,}={0,2})\b`),
	},
	{
		RuleID:      "generic-api-key",
		Description: "Generic API Key",
		Regex:       regexp.MustCompile(`(?i)(api[_-]?key|apikey|secret[_-]?key|access[_-]?token)[=:]["']?([a-zA-Z0-9_\-]{20,})["']?`),
	},
	{
		RuleID:      "password-in-url",
		Description: "Password in URL",
		Regex:       regexp.MustCompile(`(?i)(mongodb|postgresql|mysql|redis)://[^:]+:([^@]+)@`),
	},
	{
		RuleID:      "bearer-token",
		Description: "Bearer Token",
		Regex:       regexp.MustCompile(`(?i)bearer\s+([a-zA-Z0-9_\-\.]{20,})`),
	},
}
