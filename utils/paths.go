package utils

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gfunc/subconvergo/config"
)

// ResolveUnderRoot resolves target under root with symlink-safe confinement.
//
// Absolute targets and ".." components are rejected up front. Both root and
// the joined target are canonicalized with filepath.EvalSymlinks, and the
// resolved target must remain beneath the resolved root — a symlink inside
// root pointing outside it is therefore refused.
//
// The target must exist (EvalSymlinks fails otherwise); callers probing
// several candidate roots should treat an error as "try the next root".
func ResolveUnderRoot(target, root string) (string, error) {
	if target == "" {
		return "", fmt.Errorf("empty path")
	}
	if filepath.IsAbs(target) {
		return "", fmt.Errorf("absolute paths are not allowed")
	}
	for _, part := range strings.Split(filepath.ToSlash(target), "/") {
		if part == ".." {
			return "", fmt.Errorf("path traversal is not allowed")
		}
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", fmt.Errorf("failed to resolve root: %w", err)
	}
	if resolvedRoot, err := filepath.EvalSymlinks(rootAbs); err == nil {
		rootAbs = resolvedRoot
	}

	resolved, err := filepath.EvalSymlinks(filepath.Join(rootAbs, target))
	if err != nil {
		return "", fmt.Errorf("failed to resolve path: %w", err)
	}

	rel, err := filepath.Rel(rootAbs, resolved)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root directory")
	}

	return resolved, nil
}

// ResolveRulesetPath resolves a local ruleset target against the candidate
// roots (<base> and <base>/rules) with the symlink-safe confinement of
// ResolveUnderRoot and returns the first match. Absolute paths, ".."
// traversal, and symlink escapes are rejected; a target that exists under
// neither root yields "ruleset not found".
func ResolveRulesetPath(target string) (string, error) {
	basePath := config.GetBasePath()
	for _, root := range []string{basePath, filepath.Join(basePath, "rules")} {
		resolved, err := ResolveUnderRoot(target, root)
		if err != nil {
			continue
		}
		return resolved, nil
	}
	return "", fmt.Errorf("ruleset not found: %s", target)
}
