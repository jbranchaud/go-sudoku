package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var numberOfRows = 9
var numberOfColumns = 9

func main() {
	var reader io.Reader

	reader = os.Stdin

	scanner := bufio.NewScanner(reader)

	var puzzle [][]int

	for scanner.Scan() {
		row := scanner.Text()

		unparsedCells := strings.Split(row, "")
		var cells []int
		for _, unparsedCell := range unparsedCells {
			cell, err := strconv.Atoi(unparsedCell)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			cells = append(cells, cell)
		}

		puzzle = append(puzzle, cells)

		// fmt.Println(row)
	}

	_, err := validatePuzzle(puzzle)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Puzzle is valid")
	}

	printPuzzle(puzzle)
}

func validatePuzzle(puzzle [][]int) (bool, error) {
	_, err := checkForInvalidValues(puzzle)
	if err != nil {
		// early exit
		return false, err
	}

	// check each row
	for rowIndex, row := range puzzle {
		_, err := areaHasDuplicate(row, Row, rowIndex)
		if err != nil {
			return false, fmt.Errorf("Row check failed: %v", err)
		}
	}

	// check each column
	for columnIndex := range numberOfColumns {
		column := []int{}
		for _, row := range puzzle {
			column = append(column, row[columnIndex])

			_, err := areaHasDuplicate(column, Column, columnIndex)
			if err != nil {
				return false, fmt.Errorf("Column check failed: %v", err)
			}
		}
	}

	// check each 3x3 sector
	for sectorIndex := range numberOfColumns {
		sector := []int{}
		for i := range 3 {
			for j := range 3 {
				rowIndex := ((sectorIndex / 3) * 3) + i
				cellIndex := ((sectorIndex % 3) * 3) + j

				// fmt.Printf("- (%d,%d) -> %d\n", rowIndex, cellIndex, puzzle[rowIndex][cellIndex])

				sector = append(sector, puzzle[rowIndex][cellIndex])
			}
		}

		// fmt.Printf("Sector %d: %v\n", sectorIndex, sector)
		_, err := areaHasDuplicate(sector, Sector, sectorIndex)
		if err != nil {
			return false, fmt.Errorf("Sector check failed: %v", err)
		}
	}

	return true, nil
}

func checkForInvalidValues(puzzle [][]int) (bool, error) {
	for i, row := range puzzle {
		for j, cell := range row {
			if cell < 0 || cell > 9 {
				err := fmt.Errorf("Validation check failed, value '%d' at (%d,%d) is not between 0 and 9", cell, i+1, j+1)
				return false, err
			}
		}
	}

	return true, nil
}

type Area string

const (
	Row    Area = "row"
	Column Area = "column"
	Sector Area = "sector"
)

// type Address struct {
// 	Area Area
// 	X int
// 	Y int
// }

func areaHasDuplicate(cells []int, area Area, areaIndex int) (bool, error) {
	dupeIndex := hasDuplicates(cells)
	if dupeIndex >= 0 {
		var err error
		switch area {
		case Row:
			err = fmt.Errorf("Duplicate check failed, value '%d' in row %d, cell %d", cells[dupeIndex], areaIndex+1, dupeIndex+1)
		case Column:
			err = fmt.Errorf("Duplicate check failed, value '%d' in column %d, cell %d", cells[dupeIndex], areaIndex+1, dupeIndex+1)
		case Sector:
			err = fmt.Errorf("Duplicate check failed, value '%d' in sector %d, cell %d", cells[dupeIndex], areaIndex+1, dupeIndex+1)
		default:
			panic(fmt.Sprintf("Unrecognized Area '%s' provided to areaHasDuplicate", area))
		}
		return false, err
	}

	return true, nil
}

func hasDuplicates(cells []int) int {
	seen := make(map[int]bool)
	for i, cell := range cells {
		if seen[cell] {
			return i
		}
	}

	return -1
}

func removeBlanks(cells []int) []int {
	// Avoid doing this (it creates a nil slice?)
	// var compactedSlice []int

	// Instead, do this
	compactedSlice := []int{}

	for _, cell := range cells {
		if cell != 0 {
			compactedSlice = append(compactedSlice, cell)
		}
	}

	return compactedSlice
}

func printPuzzle(puzzle [][]int) {
	header :=
		"╔═══════╤═══════╤═══════╗"
	sectorDivider :=
		"╠═══════╪═══════╪═══════╣"
	footer :=
		"╚═══════╧═══════╧═══════╝"

	fmt.Println(header)
	for i, row := range puzzle {
		var builder strings.Builder
		builder.WriteString("║")
		for j, cell := range row {
			if cell == 0 {
				builder.WriteString(" _")
			} else {
				builder.WriteString(fmt.Sprintf(" %d", cell))
			}

			if j%3 == 2 {
				if j == 8 {
					builder.WriteString(" ║")
				} else {
					builder.WriteString(" │")
				}
			}
		}

		fmt.Println(builder.String())

		if i%3 == 2 && i != 8 {
			fmt.Println(sectorDivider)
		}
	}
	fmt.Println(footer)
}
