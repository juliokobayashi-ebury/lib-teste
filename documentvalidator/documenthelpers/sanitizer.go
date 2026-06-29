package documenthelpers

import "strings"

func SanitizeDocument(data string) string {
	data = strings.Replace(data, ".", "", -1)
	data = strings.Replace(data, "-", "", -1)
	data = strings.Replace(data, "/", "", -1)
	data = strings.ToUpper(data)
	data = strings.TrimSpace(data)

	return data
}