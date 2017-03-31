package utils

// StringPreview returns str shortened to max maxRunes including optional ellipsis character
func StringPreview(str string, maxRunes int) string {
	var numRunes = 0
	for index := range str {
		numRunes++
		if numRunes >= maxRunes {
			return str[:index] + "…"
		}
	}
	return str
}
