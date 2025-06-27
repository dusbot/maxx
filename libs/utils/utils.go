package utils

import "strings"

func RemoveAnyDuplicate[T comparable](list []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0)
	for _, item := range list {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	return result
}

func RemoveStrSliceDuplicate(slice []string) (result []string) {
	m := make(map[string]string, len(slice))
	for _, str := range slice {
		lowerStr := strings.ToLower(str)
		key := strings.NewReplacer(" ", "", "-", "", "_", "").Replace(lowerStr)
		m[key] = str
	}
	for _, str := range m {
		result = append(result, str)
	}
	return
}
