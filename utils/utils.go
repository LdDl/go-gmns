package utils

func TotalInts(slice []int) int {
	sum := 0
	for _, val := range slice {
		sum += val
	}
	return sum
}
