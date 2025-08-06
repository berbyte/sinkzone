package dns

import (
	"testing"
)

func TestWildcardToRegex(t *testing.T) {
	tests := []struct {
		pattern     string
		domain      string
		shouldMatch bool
	}{
		{"*github*", "github.com", true},
		{"*github*", "api.github.com", true},
		{"*github*", "githubusercontent.com", true},
		{"*github*", "example.com", false},
		{"*.example.com", "sub.example.com", true},
		{"*.example.com", "api.sub.example.com", true},
		{"*.example.com", "example.com", false},
		{"*.example.com", "other.com", false},
		{"api.*.com", "api.github.com", true},
		{"api.*.com", "api.example.com", true},
		{"api.*.com", "github.com", false},
		{"api.*.com", "api.github.org", false},
		{"exact.com", "exact.com", true},
		{"exact.com", "sub.exact.com", false},
		{"exact.com", "other.com", false},
		{"google.com", "agoogle.com", false},
	}

	for _, test := range tests {
		regex, err := wildcardToRegex(test.pattern)
		if err != nil {
			t.Errorf("Failed to compile pattern '%s': %v", test.pattern, err)
			continue
		}

		matches := regex.MatchString(test.domain)
		if matches != test.shouldMatch {
			t.Errorf("Pattern '%s' matching '%s': expected %v, got %v",
				test.pattern, test.domain, test.shouldMatch, matches)
		}
	}
}

func TestIsWildcardPattern(t *testing.T) {
	tests := []struct {
		pattern    string
		isWildcard bool
	}{
		{"*github*", true},
		{"*.example.com", true},
		{"api.*.com", true},
		{"exact.com", false},
		{"github.com", false},
		{"", false},
	}

	for _, test := range tests {
		result := isWildcardPattern(test.pattern)
		if result != test.isWildcard {
			t.Errorf("Pattern '%s': expected %v, got %v",
				test.pattern, test.isWildcard, result)
		}
	}
}
