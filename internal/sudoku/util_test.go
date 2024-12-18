package sudoku

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPop(t *testing.T) {
	t.Run("multi-item list", func(t *testing.T) {
		slice := []int{1, 2, 3}
		poppedValue, err := Pop(slice)
		assert.NoError(t, err)
		assert.Equal(t, poppedValue, 3)
	})

	t.Run("empty list", func(t *testing.T) {
		emptySlice := []int{}
		_, err := Pop(emptySlice)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "cannot pop an empty slice")
	})
}
