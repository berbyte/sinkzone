package config

import (
	"fmt"
	"os"
	"runtime"
)

// IsAdmin checks if the current process has administrative privileges
func IsAdmin() bool {
	switch runtime.GOOS {
	case "windows":
		// On Windows, check if we have the SeDebugPrivilege
		_, err := os.Open("\\\\.\\PHYSICALDRIVE0")
		return err == nil
	case "darwin", "linux":
		// On Unix-like systems, check if effective user ID is 0 (root)
		return os.Geteuid() == 0
	default:
		// For other platforms, assume no admin privileges
		return false
	}
}

// RequireAdmin checks if admin privileges are required and exits if not available
func RequireAdmin() error {
	if !IsAdmin() {
		switch runtime.GOOS {
		case "windows":
			return fmt.Errorf("administrator privileges required. Please run as Administrator")
		case "darwin", "linux":
			return fmt.Errorf("root privileges required. Please run with sudo")
		default:
			return fmt.Errorf("administrative privileges required")
		}
	}
	return nil
}

// GetAdminErrorMessage returns a platform-specific message about admin requirements
func GetAdminErrorMessage() string {
	switch runtime.GOOS {
	case "windows":
		return "Run as Administrator"
	case "darwin", "linux":
		return "Run with sudo"
	default:
		return "Run with administrative privileges"
	}
}

// IsPortPrivileged checks if a port requires admin privileges
func IsPortPrivileged(port string) bool {
	switch port {
	case "53", "80", "443":
		return true
	default:
		return false
	}
}

// CheckPortPrivileges checks if the current process can bind to the specified port
func CheckPortPrivileges(port string) error {
	if IsPortPrivileged(port) && !IsAdmin() {
		return fmt.Errorf("port %s requires %s", port, GetAdminErrorMessage())
	}
	return nil
}

// AdminError creates a platform-agnostic error message for admin privilege issues
func AdminError(err error, format string, args ...interface{}) error {
	baseErr := fmt.Errorf(format, args...)
	adminMsg := GetAdminErrorMessage()
	return fmt.Errorf("%w\nMake sure the resolver is running with %s", baseErr, adminMsg)
}
