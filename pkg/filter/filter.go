// Package filter provides domain filtering functionality for Sinkzone.
package filter

import (
	"strings"
)

// Filter represents a domain filtering engine.
type Filter struct {
	distractingSites []string
	essentialSites   []string
}

// New creates a new filter instance.
func New() *Filter {
	return &Filter{
		distractingSites: []string{
			"facebook.com", "fb.com", "instagram.com", "twitter.com", "x.com",
			"youtube.com", "youtu.be", "tiktok.com", "reddit.com", "imgur.com",
			"netflix.com", "hulu.com", "disneyplus.com", "amazon.com",
			"twitch.tv", "discord.com", "slack.com", "telegram.org",
			"snapchat.com", "pinterest.com", "linkedin.com",
		},
		essentialSites: []string{
			"google.com", "gmail.com", "github.com", "stackoverflow.com",
			"wikipedia.org", "mozilla.org", "apple.com", "microsoft.com",
			"cloudflare.com", "amazonaws.com", "digitalocean.com",
			"ubuntu.com", "debian.org", "archlinux.org",
		},
	}
}

// IsDistractingSite checks if a domain is considered distracting.
func (f *Filter) IsDistractingSite(domain string) bool {
	for _, site := range f.distractingSites {
		if strings.Contains(domain, site) {
			return true
		}
	}
	return false
}

// IsEssentialSite checks if a domain is considered essential.
func (f *Filter) IsEssentialSite(domain string) bool {
	for _, site := range f.essentialSites {
		if strings.Contains(domain, site) {
			return true
		}
	}
	return false
}

// ShouldBlock determines if a domain should be blocked based on mode and rules.
func (f *Filter) ShouldBlock(domain, mode string, customRules map[string]string) bool {
	// Check custom rules first
	if action, exists := customRules[domain]; exists {
		return action == "block"
	}

	// Apply mode-based filtering
	switch mode {
	case "off":
		return false
	case "monitor":
		return false
	case "focus":
		return f.IsDistractingSite(domain)
	case "lockdown":
		return !f.IsEssentialSite(domain)
	default:
		return false
	}
}
