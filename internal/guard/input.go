package guard

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
