package kstrings

func IsEmpty(str *string) bool {
	return str == nil || len(*str) == 0
}
