package guard

import "strings"

// Detection represents a detected secret.
type Detection struct {
	RuleID      string
	Description string
}

// Result contains the result of secret detection.
type Result struct {
	HasSecrets bool
	Detected   []Detection
}

// Sanitizer checks input for potential secrets.
type Sanitizer struct {
	rules []Rule
}

// New creates a new Sanitizer with default rules.
func New() *Sanitizer {
	return &Sanitizer{rules: DefaultRules}
}

// Check scans input for secrets and returns detection result.
func (s *Sanitizer) Check(input string) Result {
	var detected []Detection
	for _, r := range s.rules {
		if r.Regex.MatchString(input) {
			detected = append(detected, Detection{
				RuleID:      r.RuleID,
				Description: r.Description,
			})
		}
	}
	return Result{
		HasSecrets: len(detected) > 0,
		Detected:   detected,
	}
}

// SecretsError is returned when secrets are detected in input.
type SecretsError struct {
	Detected []Detection
}

func (e *SecretsError) Error() string {
	return "secrets detected: " + FormatDetections(e.Detected) +
		"\nUse --force-send to override"
}

// FormatDetections returns comma-separated list of detection descriptions.
func FormatDetections(detected []Detection) string {
	if len(detected) == 0 {
		return ""
	}
	descriptions := make([]string, len(detected))
	for i, d := range detected {
		descriptions[i] = d.Description
	}
	return strings.Join(descriptions, ", ")
}

// CheckQuery checks query for secrets and returns SecretsError if found.
// Returns nil if no secrets detected or forceSend is true.
func CheckQuery(query string, forceSend bool) error {
	sanitizer := New()
	result := sanitizer.Check(query)
	if result.HasSecrets && !forceSend {
		return &SecretsError{Detected: result.Detected}
	}
	return nil
}
