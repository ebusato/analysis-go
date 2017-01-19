package utils

func BoolToUint8(b bool) uint8 {
	if b == true {
		return 1
	}
	return 0
}
