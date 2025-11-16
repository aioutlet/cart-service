package utils

import (
	"regexp"
	"strings"
)

// GenerateVariantSKU generates a variant SKU from base SKU, color, and size
// Example: "ANT-WOM-CLO-001" + "Gray" + "M" = "ANT-WOM-CLO-001-GRAY-M"
func GenerateVariantSKU(baseSKU string, color string, size string) string {
	// Start with base SKU
	variantSKU := baseSKU
	
	// Add color if provided
	if color != "" {
		// Remove non-alphanumeric characters and convert to uppercase
		cleanColor := CleanSKUPart(color)
		if cleanColor != "" {
			variantSKU += "-" + cleanColor
		}
	}
	
	// Add size if provided
	if size != "" {
		// Remove non-alphanumeric characters and convert to uppercase
		cleanSize := CleanSKUPart(size)
		if cleanSize != "" {
			variantSKU += "-" + cleanSize
		}
	}
	
	return variantSKU
}

// CleanSKUPart cleans a string for use in SKU by removing special characters
// and converting to uppercase
func CleanSKUPart(s string) string {
	// Convert to uppercase
	s = strings.ToUpper(s)
	
	// Remove any character that's not alphanumeric
	reg := regexp.MustCompile(`[^A-Z0-9]+`)
	s = reg.ReplaceAllString(s, "")
	
	return s
}
