package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// +-----------------+-----------------+-----------------+
// | 0,0   0,1   0,2 | 0,3   0,4   0,5 | 0,6   0,7   0,8 |
// | 1,0   1,1   1,2 | 1,3   1,4   1,5 | 1,6   1,7   1,8 |
// | 2,0   2,1   2,2 | 2,3   2,4   2,5 | 2,6   2,7   2,8 |
// +-----------------+-----------------+-----------------+
// | 3,0   3,1   3,2 | 3,3   3,4   3,5 | 3,6   3,7   3,8 |
// | 4,0   4,1   4,2 | 4,3   4,4   4,5 | 4,6   4,7   4,8 |
// | 5,0   5,1   5,2 | 5,3   5,4   5,5 | 5,6   5,7   5,8 |
// +-----------------+-----------------+-----------------+
// | 6,0   6,1   6,2 | 6,3   6,4   6,5 | 6,6   6,7   6,8 |
// | 7,0   7,1   7,2 | 7,3   7,4   7,5 | 7,6   7,7   7,8 |
// | 8,0   8,1   8,2 | 8,3   8,4   8,5 | 8,6   8,7   8,8 |
// +-----------------+-----------------+-----------------+

func TestGetSectorNumberForCell(t *testing.T) {
	t.Run("first sector", func(t *testing.T) {
		firstA := GetSectorNumberForCell(0, 0)
		assert.Equal(t, firstA, 0)

		firstB := GetSectorNumberForCell(2, 2)
		assert.Equal(t, firstB, 0)

		firstC := GetSectorNumberForCell(2, 0)
		assert.Equal(t, firstC, 0)
	})

	t.Run("second sector", func(t *testing.T) {
		secondA := GetSectorNumberForCell(0, 3)
		assert.Equal(t, secondA, 1)

		secondB := GetSectorNumberForCell(2, 3)
		assert.Equal(t, secondB, 1)

		secondC := GetSectorNumberForCell(2, 5)
		assert.Equal(t, secondC, 1)
	})

	t.Run("ninth sector", func(t *testing.T) {
		ninthA := GetSectorNumberForCell(6, 8)
		assert.Equal(t, ninthA, 8)

		ninthB := GetSectorNumberForCell(7, 7)
		assert.Equal(t, ninthB, 8)

		ninthC := GetSectorNumberForCell(8, 6)
		assert.Equal(t, ninthC, 8)
	})
}
