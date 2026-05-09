package pkg

func TruncateString(message string, limit int) string {
	if len(message) <= limit {
		return message
	}

	runes := []rune(message)
	if len(runes) <= limit {
		return message
	}

	return string(runes[:limit]) + "..."
}
