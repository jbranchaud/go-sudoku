package main

import "fmt"

func zero[T any]() T {
	return *new(T)
}

func Pop[T any](slice []T) (T, error) {
	if len(slice) == 0 {
		return zero[T](), fmt.Errorf("cannot pop an empty slice")
	}

	lastItem := slice[len(slice)-1]

	slice = slice[:len(slice)-1]

	return lastItem, nil
}

func GetSectorNumberForCell(row int, cell int) int {
	offset := row / 3 * 3
	return offset + (cell / 3)
}
