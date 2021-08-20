package fileutils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/scanner"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

// PatternMatcher allows checking paths against a list of patterns
type PatternMatcher struct {
	patterns   []*Pattern
	exclusions bool
}

// NewPatternMatcher creates a new matcher object for specific patterns that can
// be used later to match against patterns against paths
func NewPatternMatcher(patterns []string) (*PatternMatcher, error) {
	pm := &PatternMatcher{
		patterns: make([]*Pattern, 0, len(patterns)),
	}

	for _, p := range patterns {
		// Eliminate leading and trailing whitespace.
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		p = filepath.Clean(p)
		newp := &Pattern{}

		if p[0] == '!' {
			if len(p) == 1 {
				return nil, errors.New("illegal exclusion pattern: \"!\"")
			}

			newp.exclusion = true
			p = strings.TrimPrefix(filepath.Clean(p[1:]), "/")
			pm.exclusions = true
		}

		if _, err := filepath.Match(p, "."); err != nil {
			return nil, err
		}

		newp.cleanedPattern = p
		newp.dirs = strings.Split(p, string(os.PathSeparator))
		pm.patterns = append(pm.patterns, newp)
	}

	return pm, nil
}

func (pm *PatternMatcher) Matches(file string) (bool, error) {
	matched := false
	file = filepath.FromSlash(file)

	for _, pattern := range pm.patterns {
		negative := false

		if pattern.exclusion {
			negative = true
		}

		match, err := pattern.match(file)
		if err != nil {
			return false, err
		}

		if match {
			matched = !negative
		}
	}

	if matched {
		logrus.Debugf("Skipping excluded path: %s", file)
	}

	return matched, nil
}

type MatchResult struct {
	isMatched         bool
	matches, excludes uint
}

// Excludes returns true if the overall result is matched
func (m *MatchResult) IsMatched() bool {
	return m.isMatched
}

// Excludes returns the amount of matches of an MatchResult
func (m *MatchResult) Matches() uint {
	return m.matches
}

// Excludes returns the amount of excludes of an MatchResult
func (m *MatchResult) Excludes() uint {
	return m.excludes
}

// MatchesResult verifies the provided filepath against all patterns.
// It returns the `*MatchResult` result for the patterns on success, otherwise
// an error. This method is not safe to be called concurrently.
func (pm *PatternMatcher) MatchesResult(file string) (res *MatchResult, err error) {
	file = filepath.FromSlash(file)
	res = &MatchResult{false, 0, 0}

	for _, pattern := range pm.patterns {
		negative := false

		if pattern.exclusion {
			negative = true
		}

		match, err := pattern.match(file)
		if err != nil {
			return nil, err
		}

		if match {
			res.isMatched = !negative
			if negative {
				res.excludes++
			} else {
				res.matches++
			}
		}
	}

	if res.matches > 0 {
		logrus.Debugf("Skipping excluded path: %s", file)
	}

	return res, nil
}
