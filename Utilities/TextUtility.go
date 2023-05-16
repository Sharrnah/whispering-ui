package Utilities

func CountUTF16CodeUnits(s string) int {
	count := 0
	for _, r := range s {
		if r >= 0x10000 {
			count += 2
		} else {
			count++
		}
	}
	return count
}
