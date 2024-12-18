package main

import (
	"cmp"
	"math/rand"
)

func GetSectorNumberForCell(row int, cell int) int {
	offset := row / 3 * 3
	return offset + (cell / 3)
}

// Fisher-Yates shuffle
func Shuffle[T cmp.Ordered](slice []T, rng *rand.Rand) ([]T, error) {
	// this bit seems odd, I ought to be able to guarantee that an rng has been passed in
	var internalRng rand.Rand
	if rng == nil {
		internalRng = rand.Rand{}
	} else {
		internalRng = *rng
	}

	if len(slice) == 0 {
		return slice, nil
	}

	for i := range len(slice) - 1 {
		j := internalRng.Intn(len(slice) - 1)
		slice[i], slice[j] = slice[j], slice[i]
	}

	return slice, nil
}
