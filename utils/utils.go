package utils

var (
	Green = string([]byte{27, 91, 57, 55, 59, 52, 50, 109})
	Red   = string([]byte{27, 91, 57, 55, 59, 52, 49, 109})
	Reset = string([]byte{27, 91, 48, 109})
	White = string([]byte{27, 91, 57, 48, 59, 52, 55, 109})
)

func StringDefault(orig, val string) string {
	if orig == "" {
		return val
	} else {
		return orig
	}
}

func FindStringInArray(item string, items []string) bool {
	for _, i := range items {
		if i == item {
			return true
		}
	}
	return false
}
