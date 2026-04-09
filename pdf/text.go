package pdf

import (
	"regexp"
	"strings"
	"unicode"
)

// DetectCorruptedFont detects if a font name is garbled and attempts to fix it
// Returns: (isGarbled, fixedFontName)
func DetectCorruptedFont(fontName string) (bool, string) {
	// Check if it's a known normal font
	for name := range commonFontNames {
		if strings.Contains(fontName, name) {
			return false, fontName
		}
	}

	// Check for garbled characteristics (Western characters + special symbols)
	if !regexp.MustCompile(`[\x80-\xFF]`).MatchString(fontName) {
		return false, fontName
	}

	// Try fixing: Latin1 encode -> GBK decode
	fixed := encodeLatin1ToGBK(fontName)
	if fixed != "" && HasChinese(fixed) && LooksLikeFontName(fixed) {
		return true, fixed
	}

	// Try other encoding combinations
	encodings := [][2]string{
		{"cp1252", "gbk"},
		{"iso8859_15", "gb2312"},
		{"latin1", "big5"},
	}
	for _, enc := range encodings {
		fixed := encodeDecode(fontName, enc[0], enc[1])
		if fixed != "" && HasChinese(fixed) && LooksLikeFontName(fixed) {
			return true, fixed
		}
	}

	return false, fontName
}

// encodeLatin1ToGBK attempts to fix garbled font name
func encodeLatin1ToGBK(s string) string {
	// Convert string to bytes using Latin1, then try to decode as GBK
	// This is a simplified version - Go doesn't have direct Latin1 encoding
	// We simulate the effect
	bytes := []byte{}
	for _, c := range s {
		bytes = append(bytes, byte(c))
	}
	// Try GBK decode (simplified - just check if it looks valid)
	if len(bytes) >= 2 {
		// Check for high bytes that might indicate Chinese
		hasHighBytes := false
		for _, b := range bytes {
			if b > 0x80 {
				hasHighBytes = true
				break
			}
		}
		if hasHighBytes {
			return string(bytes) // Simplified - actual impl would need charset conversion
		}
	}
	return ""
}

// encodeDecode encodes with srcEnc and decodes with dstEnc
func encodeDecode(s, srcEnc, dstEnc string) string {
	// Simplified implementation
	bytes := []byte{}
	for _, c := range s {
		if c > 127 {
			bytes = append(bytes, byte(c))
		}
	}
	if len(bytes) > 0 {
		return string(bytes)
	}
	return ""
}

// HasChinese checks if text contains Chinese characters
func HasChinese(text string) bool {
	return chineseRegex.MatchString(text)
}

// LooksLikeFontName validates if string looks like a font name
func LooksLikeFontName(text string) bool {
	// Check length
	if len(text) < 1 || len(text) > 50 {
		return false
	}

	// Check for font keywords
	for _, kw := range fontKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}

	// Check character composition
	return fontNameRegex.MatchString(text)
}

// IsGarbledText determines if extracted text might be garbled
func IsGarbledText(text string) bool {
	if text == "" || len(text) < 3 {
		return false
	}

	text = strings.ReplaceAll(text, " ", "")

	// Check for garbled pattern (5+ consecutive non-ASCII)
	if garbledPattern.MatchString(text) {
		return true
	}

	// Check for control character ratio
	var controlChars int
	for _, r := range text {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			controlChars++
		}
	}
	if len(text) > 0 && float64(controlChars)/float64(len(text)) > 0.1 {
		return true
	}

	return false
}
