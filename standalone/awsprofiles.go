package main

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	credSectionRe   = regexp.MustCompile(`^\[([^\]]+)\]$`)
	configProfileRe = regexp.MustCompile(`^\[profile\s+([^\]]+)\]$`)
)

// loadAwsProfiles reads profile names from ~/.aws/credentials and ~/.aws/config
// and returns a sorted, de-duplicated list with "default" first if present.
func loadAwsProfiles() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return []string{"default"}
	}

	set := map[string]struct{}{}

	// ~/.aws/credentials: sections are plain profile names, e.g. [myprofile] or [default]
	scanFile(filepath.Join(home, ".aws", "credentials"), func(line string) {
		if m := credSectionRe.FindStringSubmatch(line); m != nil {
			set[m[1]] = struct{}{}
		}
	})

	// ~/.aws/config: sections are "[profile myprofile]" except "[default]"
	scanFile(filepath.Join(home, ".aws", "config"), func(line string) {
		if m := configProfileRe.FindStringSubmatch(line); m != nil {
			set[m[1]] = struct{}{}
		} else if m := credSectionRe.FindStringSubmatch(line); m != nil && m[1] == "default" {
			set[m[1]] = struct{}{}
		}
	})

	if len(set) == 0 {
		return []string{"default"}
	}

	profiles := make([]string, 0, len(set))
	for p := range set {
		profiles = append(profiles, p)
	}
	sort.Strings(profiles)

	// Put "default" first if present
	for i, p := range profiles {
		if p == "default" {
			profiles = append(profiles[:i], profiles[i+1:]...)
			profiles = append([]string{"default"}, profiles...)
			break
		}
	}
	return profiles
}

// scanFile reads a file line by line, ignoring commented-out lines (starting with '#' or ';'),
// and invokes fn on each trimmed line. Missing files are silently ignored.
func scanFile(path string, fn func(line string)) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		fn(line)
	}
}
