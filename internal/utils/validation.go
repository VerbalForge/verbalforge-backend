package utils

import "strings"

// Contains checks if a string contains a substring
func Contains(str, substr string) bool {
	return strings.Contains(str, substr)
}

// IsEmail checks if a string is an email address
func IsEmail(str string) bool {
	return Contains(str, "@")
}

// TrimSpace removes leading and trailing whitespace
func TrimSpace(str string) string {
	return strings.TrimSpace(str)
}
