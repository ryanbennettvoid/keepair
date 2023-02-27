package partition

func GenerateDeterministicPartitionKey(key string, numPartitions int) int {
	n := 0
	for _, char := range []byte(key) {
		n += int(char)
	}
	return n % numPartitions
}
