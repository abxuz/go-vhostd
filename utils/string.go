package utils

import "strings"

func ExistEmptyString(trim bool, ss ...string) bool {
	for _, s := range ss {
		if trim {
			s = strings.TrimSpace(s)
		}
		if s == "" {
			return true
		}
	}
	return false
}
