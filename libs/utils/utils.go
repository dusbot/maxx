package utils

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
