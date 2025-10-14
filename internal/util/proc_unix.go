//go:build linux

package util

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ReadProcUID returns real UID of a PID using /proc/<pid>/status.
func ReadProcUID(pid int) (string, bool) {
	p := filepath.Join("/proc", strconv.Itoa(pid), "status")
	f, err := os.Open(p)
	if err != nil {
		return "", false
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return fields[1], true // real UID
			}
			break
		}
	}
	return "", false
}

// BuildUIDMap parses /etc/passwd and returns uid->username.
func BuildUIDMap() map[string]string {
	m := make(map[string]string)
	f, err := os.Open("/etc/passwd")
	if err != nil {
		return m
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.Split(line, ":")
		if len(parts) < 3 {
			continue
		}
		name := parts[0]
		uid := parts[2]
		m[uid] = name
	}
	return m
}
