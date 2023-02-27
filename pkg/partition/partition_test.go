package partition

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"keepair/pkg/common"

	"github.com/stretchr/testify/assert"
)

type Stats struct {
	Index0 int
	Index1 int
	Index2 int
	Index3 int
}

// TestGenerateDeterministicPartitionKey checks that the function
// generates roughly evenly distributed indexes for 0 to N,
// with N being the number of partitions
func TestGenerateDeterministicPartitionKey(t *testing.T) {

	numPartitions := 4
	numKeys := 1000

	stats := Stats{
		Index0: 0,
		Index1: 0,
		Index2: 0,
		Index3: 0,
	}

	for i := 0; i < numKeys; i++ {
		numChars := (rand.Int() % 20) + 1
		key := common.GenerateRandomString(numChars)
		index := GenerateDeterministicPartitionKey(key, numPartitions)
		switch index {
		case 0:
			stats.Index0++
		case 1:
			stats.Index1++
		case 2:
			stats.Index2++
		case 3:
			stats.Index3++
		default:
			panic(fmt.Errorf("invalid index: %d", index))
		}
	}

	perfectPartitionSize := numKeys / numPartitions
	maxMarginOfError := float64(perfectPartitionSize) * 0.1 // 10%
	for _, actualPartitionSize := range []int{stats.Index0, stats.Index1, stats.Index2, stats.Index3} {
		marginOfError := math.Abs(float64(actualPartitionSize) - float64(perfectPartitionSize))
		assert.Truef(t, marginOfError < maxMarginOfError, "max margin of error exceeded: %f/%f", marginOfError, maxMarginOfError)
	}

}
