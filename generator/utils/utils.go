package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
)

// GetFieldTag gets the tag value of a struct field
func GetFieldTag(tagType, tagName string, s interface{}, defaultTag string) string {
	t := reflect.TypeOf(s).Elem()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Name == tagName {
			tag := field.Tag.Get(tagType)
			if tag != "" {
				parts := strings.Split(tag, ",")
				return parts[0]
			}
		}
	}
	return defaultTag
}

// FetchRuleset fetches a ruleset from a URL or local file
func FetchRuleset(path string) (string, error) {
	// Check if it's a URL
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		resp, err := http.Get(path)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}
		return string(body), nil
	}

	// Check local file
	// Try absolute path first
	if _, err := os.Stat(path); err == nil {
		content, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	// Try relative to base path
	basePath := config.Global.Common.BasePath
	fullPath := filepath.Join(basePath, path)
	if _, err := os.Stat(fullPath); err == nil {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	// Try relative to base/rules
	fullPath = filepath.Join(basePath, "rules", path)
	if _, err := os.Stat(fullPath); err == nil {
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	return "", fmt.Errorf("ruleset not found: %s", path)
}

// FilterProxiesByRules filters proxies based on rule patterns
func FilterProxiesByRules(proxies []core.ProxyInterface, rules []string) []string {
	var result []string
	seen := make(map[string]bool)

	// Handle empty rules -> select all
	if len(rules) == 0 {
		for _, p := range proxies {
			result = append(result, p.GetRemark())
		}
		return result
	}

	for _, rule := range rules {
		if rule == "" {
			continue
		}

		if strings.HasPrefix(rule, "[]") {
			name := strings.TrimPrefix(rule, "[]")
			if !seen[name] {
				result = append(result, name)
				seen[name] = true
			}
			continue
		}

		// Regex or Special Matcher
		matcher := rule
		isSpecial := strings.HasPrefix(matcher, "!!")
		var re *regexp.Regexp

		if !isSpecial {
			// Treat as regex
			re, _ = regexp.Compile(matcher)
		}

		for _, p := range proxies {
			name := p.GetRemark()
			if seen[name] {
				continue
			}

			matched := false
			if isSpecial {
				matched, _ = ApplyMatcher(matcher, p)
			} else if re != nil {
				matched = re.MatchString(name)
			} else {
				// Fallback to substring if regex compilation failed
				matched = strings.Contains(name, matcher)
			}

			if matched {
				result = append(result, name)
				seen[name] = true
			}
		}
	}

	return result
}

// ApplyMatcher applies special matchers
func ApplyMatcher(rule string, p core.ProxyInterface) (bool, string) {
	// Handle !!GROUP= matcher
	if strings.HasPrefix(rule, "!!GROUP=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			groupPattern := strings.TrimPrefix(parts[1], "GROUP=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			matched, _ := regexp.MatchString(groupPattern, p.GetGroup())
			return matched, realRule
		}
	}

	// Handle !!TYPE= matcher
	if strings.HasPrefix(rule, "!!TYPE=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			typePattern := strings.TrimPrefix(parts[1], "TYPE=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			proxyType := strings.ToUpper(p.GetType())
			matched, _ := regexp.MatchString("(?i)^("+typePattern+")$", proxyType)
			return matched, realRule
		}
	}

	// Handle !!PORT= matcher
	if strings.HasPrefix(rule, "!!PORT=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			portPattern := strings.TrimPrefix(parts[1], "PORT=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			matched := MatchRange(portPattern, p.GetPort())
			return matched, realRule
		}
	}

	// Handle !!SERVER= matcher
	if strings.HasPrefix(rule, "!!SERVER=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			serverPattern := strings.TrimPrefix(parts[1], "SERVER=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			matched, _ := regexp.MatchString(serverPattern, p.GetServer())
			return matched, realRule
		}
	}

	// Handle !!GROUPID= matcher
	if strings.HasPrefix(rule, "!!GROUPID=") {
		parts := strings.SplitN(rule, "!!", 3)
		if len(parts) >= 2 {
			idPattern := strings.TrimPrefix(parts[1], "GROUPID=")
			realRule := ""
			if len(parts) > 2 {
				realRule = parts[2]
			}
			matched := MatchRange(idPattern, p.GetGroupId())
			return matched, realRule
		}
	}

	// No special matcher, return rule as-is
	return true, rule
}

// MatchRange checks if a value matches a range pattern
func MatchRange(pattern string, value int) bool {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return true
	}

	// Handle comma-separated values: "1,2,3"
	if strings.Contains(pattern, ",") {
		parts := strings.Split(pattern, ",")
		for _, part := range parts {
			if MatchRange(strings.TrimSpace(part), value) {
				return true
			}
		}
		return false
	}

	// Handle ranges: "1-5"
	if strings.Contains(pattern, "-") {
		parts := strings.Split(pattern, "-")
		if len(parts) == 2 {
			start, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			end, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				return value >= start && value <= end
			}
		}
	}

	// Handle single value: "5"
	if num, err := strconv.Atoi(pattern); err == nil {
		return value == num
	}

	return false
}
