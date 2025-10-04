package output

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var (
	// Color definitions
	promptColor    = color.New(color.FgCyan, color.Bold)
	userColor      = color.New(color.FgGreen)
	assistantColor = color.New(color.FgWhite)
	errorColor     = color.New(color.FgRed, color.Bold)
	warningColor   = color.New(color.FgYellow)
	successColor   = color.New(color.FgGreen, color.Bold)
	infoColor      = color.New(color.FgBlue)
	commandColor   = color.New(color.FgMagenta)

	// Secret patterns to filter
	secretPatterns = []*regexp.Regexp{
		// Base64 encoded secrets (common in Kubernetes)
		regexp.MustCompile(`"data"\s*:\s*\{[^}]*"[^"]*"\s*:\s*"[A-Za-z0-9+/]{20,}={0,2}"`),
		regexp.MustCompile(`(?i)"password"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"token"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"secret"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"key"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"apikey"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"api_key"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"access_key"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"secret_key"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"private_key"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`(?i)"client_secret"\s*:\s*"[^"]+"`),
		// Password without quotes
		regexp.MustCompile(`(?i)Password:\s*"[^"]+"`),
		// Certificate data
		regexp.MustCompile(`"ca\.crt"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`"tls\.crt"\s*:\s*"[^"]+"`),
		regexp.MustCompile(`"tls\.key"\s*:\s*"[^"]+"`),
		// Bearer tokens
		regexp.MustCompile(`Bearer\s+[A-Za-z0-9\-._~+/]+`),
		// Generic base64 patterns that look like secrets
		// Must contain base64 characters (+/) and optional padding (=)
		// This avoids matching DNS names which contain hyphens and dots
		regexp.MustCompile(`\b[A-Za-z0-9+/]{40,}={0,2}\b`),
		// Base64 with mandatory padding that indicates encoded data
		regexp.MustCompile(`\b[A-Za-z0-9+/]{32,}==?\b`),
	}

	// Replacement text for filtered secrets
	secretReplacement = "***REDACTED***"
)

// Printer provides colored and filtered output for the console
type Printer struct {
	filterSecrets bool
}

// NewPrinter creates a new printer instance
func NewPrinter(filterSecrets bool) *Printer {
	return &Printer{
		filterSecrets: filterSecrets,
	}
}

// PrintPrompt prints the console prompt with color
func (p *Printer) PrintPrompt() {
	_, _ = promptColor.Print("> ")
}

// PrintUser prints user input with appropriate color
func (p *Printer) PrintUser(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = userColor.Print(output)
}

// PrintUserln prints user input with newline
func (p *Printer) PrintUserln(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = userColor.Println(output)
}

// PrintAssistant prints assistant output with appropriate color
func (p *Printer) PrintAssistant(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = assistantColor.Print(output)
}

// PrintAssistantln prints assistant output with newline
func (p *Printer) PrintAssistantln(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = assistantColor.Println(output)
}

// PrintError prints error messages
func (p *Printer) PrintError(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = errorColor.Print(output)
}

// PrintErrorln prints error messages with newline
func (p *Printer) PrintErrorln(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = errorColor.Println(output)
}

// PrintWarning prints warning messages
func (p *Printer) PrintWarning(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = warningColor.Print(output)
}

// PrintWarningln prints warning messages with newline
func (p *Printer) PrintWarningln(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = warningColor.Println(output)
}

// PrintSuccess prints success messages
func (p *Printer) PrintSuccess(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = successColor.Print(output)
}

// PrintSuccessln prints success messages with newline
func (p *Printer) PrintSuccessln(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = successColor.Println(output)
}

// PrintInfo prints informational messages
func (p *Printer) PrintInfo(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = infoColor.Print(output)
}

// PrintInfoln prints informational messages with newline
func (p *Printer) PrintInfoln(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = infoColor.Println(output)
}

// PrintCommand prints command output
func (p *Printer) PrintCommand(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = commandColor.Print(output)
}

// PrintCommandln prints command output with newline
func (p *Printer) PrintCommandln(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	_, _ = commandColor.Println(output)
}

// Print prints regular output (no color)
func (p *Printer) Print(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	fmt.Print(output)
}

// Println prints regular output with newline (no color)
func (p *Printer) Println(format string, a ...interface{}) {
	output := fmt.Sprintf(format, a...)
	output = p.filterOutput(output)
	fmt.Println(output)
}

// filterOutput filters sensitive information from the output
func (p *Printer) filterOutput(output string) string {
	if !p.filterSecrets {
		return output
	}

	filtered := output

	// Apply all secret patterns
	for _, pattern := range secretPatterns {
		filtered = pattern.ReplaceAllString(filtered, secretReplacement)
	}

	// Additional filtering for Kubernetes secret objects
	if strings.Contains(filtered, "kind: Secret") || strings.Contains(filtered, "\"kind\":\"Secret\"") {
		filtered = p.filterKubernetesSecrets(filtered)
	}

	return filtered
}

// filterKubernetesSecrets provides additional filtering for Kubernetes secret objects
func (p *Printer) filterKubernetesSecrets(output string) string {
	// Look for data fields in YAML format
	yamlDataPattern := regexp.MustCompile(`(?m)^data:\s*\n((?:\s+\w+:.*\n?)+)`)
	output = yamlDataPattern.ReplaceAllStringFunc(output, func(match string) string {
		lines := strings.Split(match, "\n")
		result := []string{lines[0]} // Keep "data:" line
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				parts := strings.SplitN(lines[i], ":", 2)
				if len(parts) == 2 {
					result = append(result, fmt.Sprintf("%s: %s", parts[0], secretReplacement))
				} else {
					result = append(result, lines[i])
				}
			} else if i < len(lines)-1 {
				// Keep empty lines except the last one
				result = append(result, lines[i])
			}
		}
		return strings.Join(result, "\n")
	})

	// Look for stringData fields in YAML format
	yamlStringDataPattern := regexp.MustCompile(`(?m)^stringData:\s*\n((?:\s+\w+:.*\n)+)`)
	output = yamlStringDataPattern.ReplaceAllStringFunc(output, func(match string) string {
		lines := strings.Split(match, "\n")
		result := []string{lines[0]} // Keep "stringData:" line
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) != "" {
				parts := strings.SplitN(lines[i], ":", 2)
				if len(parts) == 2 {
					result = append(result, fmt.Sprintf("%s: %s", parts[0], secretReplacement))
				} else {
					result = append(result, lines[i])
				}
			}
		}
		return strings.Join(result, "\n")
	})

	return output
}

// FilterSecrets filters sensitive information from a string
func FilterSecrets(input string) string {
	p := &Printer{filterSecrets: true}
	return p.filterOutput(input)
}

// IsLikelySecret checks if a value looks like a secret
func IsLikelySecret(value string) bool {
	// Check if it looks like a JWT token first (3 parts separated by dots)
	if strings.Count(value, ".") == 2 && len(value) > 50 {
		return true
	}

	// Check if it's base64 encoded and has reasonable length
	// Must check for = padding at the end for valid base64
	if len(value) >= 20 && regexp.MustCompile(`^[A-Za-z0-9+/]+={0,2}$`).MatchString(value) {
		if _, err := base64.StdEncoding.DecodeString(value); err == nil {
			return true
		}
	}

	// Check for common secret patterns
	lowerValue := strings.ToLower(value)
	secretKeywords := []string{
		"password", "passwd", "secret", "token", "key", "apikey",
		"api_key", "access_key", "secret_key", "private_key",
		"client_secret", "bearer",
	}

	for _, keyword := range secretKeywords {
		if strings.Contains(lowerValue, keyword) {
			return true
		}
	}

	return false
}
