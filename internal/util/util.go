package util

import "fmt"

// InterfaceToString converts a interface{} variable to a string.
func InterfaceToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case []byte:
		return string(v)
	case error:
		return v.Error()
	}
	return fmt.Sprintf("%v", value)
}
