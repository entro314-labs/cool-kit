package ui

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// DeploymentError represents a structured deployment error with context
type DeploymentError struct {
	Provider   string
	Operation  string
	Code       string
	Message    string
	Suggestion string
	Raw        error
}

func (e *DeploymentError) Error() string {
	return e.Message
}

func (e *DeploymentError) Unwrap() error {
	return e.Raw
}

// NewDeploymentError creates a structured deployment error
func NewDeploymentError(provider, operation string, cause error) *DeploymentError {
	de := &DeploymentError{
		Provider:  provider,
		Operation: operation,
		Raw:       cause,
	}
	de.parse()
	return de
}

// parse extracts structured info from the raw error
func (e *DeploymentError) parse() {
	if e.Raw == nil {
		e.Message = "Unknown error"
		return
	}

	errStr := e.Raw.Error()

	// Try to extract Azure API error
	if azureErr := extractAzureError(errStr); azureErr != nil {
		e.Code = azureErr.Code
		e.Message = azureErr.Message
		e.Suggestion = azureErr.Suggestion
		return
	}

	// Try to extract AWS error
	if awsErr := extractAWSError(errStr); awsErr != nil {
		e.Code = awsErr.Code
		e.Message = awsErr.Message
		e.Suggestion = awsErr.Suggestion
		return
	}

	// Try to extract GCP error
	if gcpErr := extractGCPError(errStr); gcpErr != nil {
		e.Code = gcpErr.Code
		e.Message = gcpErr.Message
		e.Suggestion = gcpErr.Suggestion
		return
	}

	// Fallback: clean up raw error
	e.Message = cleanErrorMessage(errStr)
}

type parsedError struct {
	Code       string
	Message    string
	Suggestion string
}

// extractAzureError parses Azure API error responses
func extractAzureError(errStr string) *parsedError {
	// Look for Azure error JSON pattern
	jsonRe := regexp.MustCompile(`\{[^{}]*"error"\s*:\s*\{[^{}]*"code"\s*:\s*"([^"]+)"[^{}]*"message"\s*:\s*"([^"]+)"`)
	matches := jsonRe.FindStringSubmatch(errStr)
	if len(matches) >= 3 {
		code := matches[1]
		msg := matches[2]
		return &parsedError{
			Code:       code,
			Message:    msg,
			Suggestion: getAzureSuggestion(code),
		}
	}

	// Try simpler patterns
	if strings.Contains(errStr, "PropertyChangeNotAllowed") {
		return &parsedError{
			Code:       "PropertyChangeNotAllowed",
			Message:    "Cannot modify existing VM configuration",
			Suggestion: "Delete the resource group first: az group delete --name coolify-rg --yes",
		}
	}

	if strings.Contains(errStr, "ResourceGroupNotFound") {
		return &parsedError{
			Code:       "ResourceGroupNotFound",
			Message:    "Resource group does not exist",
			Suggestion: "The resource group will be created automatically",
		}
	}

	if strings.Contains(errStr, "AuthorizationFailed") {
		return &parsedError{
			Code:       "AuthorizationFailed",
			Message:    "Insufficient permissions",
			Suggestion: "Ensure your Azure account has Contributor role on the subscription",
		}
	}

	return nil
}

func getAzureSuggestion(code string) string {
	suggestions := map[string]string{
		"PropertyChangeNotAllowed":  "Delete existing resources first: az group delete --name coolify-rg --yes",
		"ResourceNotFound":          "The resource may have been deleted or doesn't exist",
		"AuthorizationFailed":       "Check your Azure permissions - Contributor role required",
		"QuotaExceeded":             "Request a quota increase in Azure portal or try a different region",
		"SkuNotAvailable":           "This VM size is not available in the region. Try a different VM size or region",
		"OperationNotAllowed":       "This operation is not permitted. Check subscription restrictions",
		"SubscriptionNotRegistered": "Register the resource provider: az provider register --namespace Microsoft.Compute",
		"InvalidParameter":          "Check your configuration values",
		"ConflictingUserInput":      "There's a conflict with existing resources. Try deleting and recreating",
	}
	if s, ok := suggestions[code]; ok {
		return s
	}
	return ""
}

// extractAWSError parses AWS error responses
func extractAWSError(errStr string) *parsedError {
	if strings.Contains(errStr, "UnauthorizedAccess") || strings.Contains(errStr, "AccessDenied") {
		return &parsedError{
			Code:       "AccessDenied",
			Message:    "AWS access denied",
			Suggestion: "Check your AWS credentials: aws configure list",
		}
	}
	if strings.Contains(errStr, "InstanceLimitExceeded") {
		return &parsedError{
			Code:       "InstanceLimitExceeded",
			Message:    "Instance limit exceeded",
			Suggestion: "Request a limit increase or terminate unused instances",
		}
	}
	return nil
}

// extractGCPError parses GCP error responses
func extractGCPError(errStr string) *parsedError {
	if strings.Contains(errStr, "PERMISSION_DENIED") {
		return &parsedError{
			Code:       "PERMISSION_DENIED",
			Message:    "GCP permission denied",
			Suggestion: "Run: gcloud auth login && gcloud config set project YOUR_PROJECT",
		}
	}
	return nil
}

// cleanErrorMessage removes noise from error strings
func cleanErrorMessage(errStr string) string {
	// Remove HTTP response noise
	patterns := []string{
		`PUT https://[^\s]+`,
		`GET https://[^\s]+`,
		`POST https://[^\s]+`,
		`DELETE https://[^\s]+`,
		`-{10,}`,
		`RESPONSE \d+: \d+ [A-Za-z]+`,
		`ERROR CODE: [A-Za-z]+`,
	}

	result := errStr
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		result = re.ReplaceAllString(result, "")
	}

	// Clean up whitespace
	result = strings.TrimSpace(result)
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")

	// Limit length
	if len(result) > 200 {
		result = result[:200] + "..."
	}

	return result
}

// FormatError formats any error for user display
func FormatError(err error) string {
	if err == nil {
		return ""
	}

	// Check if it's already a DeploymentError
	if de, ok := err.(*DeploymentError); ok {
		return formatDeploymentError(de)
	}

	// Parse and format
	de := NewDeploymentError("", "", err)
	return formatDeploymentError(de)
}

func formatDeploymentError(de *DeploymentError) string {
	var b strings.Builder

	// Error message
	if de.Code != "" {
		b.WriteString(fmt.Sprintf("[%s] %s", de.Code, de.Message))
	} else {
		b.WriteString(de.Message)
	}

	// Suggestion if available
	if de.Suggestion != "" {
		b.WriteString("\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FBCA04")).Render("💡 Fix: "))
		b.WriteString(de.Suggestion)
	}

	return b.String()
}

// FormatErrorBox renders an error in a styled box
func FormatErrorBox(title string, err error, width int) string {
	formattedErr := FormatError(err)

	// Word wrap the error message
	wrapped := wordWrap(formattedErr, width-8)

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF5F87")).
		Padding(0, 2).
		Width(width - 4)

	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5F87")).
		Bold(true)

	content := titleStyle.Render("❌ "+title) + "\n\n" + wrapped

	return boxStyle.Render(content)
}

// wordWrap wraps text to a maximum width
func wordWrap(text string, width int) string {
	if width <= 0 {
		width = 80
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for i, line := range lines {
		if i > 0 {
			result.WriteString("\n")
		}

		words := strings.Fields(line)
		if len(words) == 0 {
			continue
		}

		currentLine := words[0]
		for _, word := range words[1:] {
			if len(currentLine)+1+len(word) > width {
				result.WriteString(currentLine)
				result.WriteString("\n")
				currentLine = word
			} else {
				currentLine += " " + word
			}
		}
		result.WriteString(currentLine)
	}

	return result.String()
}

// TryParseJSON attempts to pretty-print JSON, returns original if not JSON
func TryParseJSON(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "{") && !strings.HasPrefix(s, "[") {
		return s
	}

	var data interface{}
	if err := json.Unmarshal([]byte(s), &data); err != nil {
		return s
	}

	pretty, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return s
	}

	return string(pretty)
}
