package db

func fixIsoDateStr(s string) string {
	const isoLen = len("2006-01-02T15:04:05Z")

	if len(s) == isoLen && s[4] == '-' && s[7] == '-' && s[10] == 'T' && s[isoLen-1] == 'Z' {
		b := []byte(s[:isoLen-1])
		b[10] = ' '
		return string(b)
	}

	return s
}
