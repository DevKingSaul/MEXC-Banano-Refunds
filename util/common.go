package util

func IsZero(bytes []byte) bool {
	for _, b := range bytes {
		if b != 0 {
			return false
		}
	}

	return true
}