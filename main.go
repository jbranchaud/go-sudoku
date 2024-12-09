package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

var gridSize = 9

type Puzzle struct {
	Board [][]int
}

func (puz *Puzzle) getRow(rowIndex int) []int {
	if rowIndex < 0 || rowIndex > 8 {
		panic(fmt.Sprintf("Invalid rowIndex %d", rowIndex))
	}

	return puz.Board[rowIndex]
}

func (puz *Puzzle) getColumn(colIndex int) []int {
	column := []int{}
	for _, row := range puz.Board {
		column = append(column, row[colIndex])
	}

	return column
}

func (puz *Puzzle) getSector(secIndex int) []int {
	sector := []int{}
	for i := range 3 {
		for j := range 3 {
			rowIndex := ((secIndex / 3) * 3) + i
			cellIndex := ((secIndex % 3) * 3) + j

			sector = append(sector, puz.Board[rowIndex][cellIndex])
		}
	}

	return sector
}

func main() {
	var reader io.Reader

	reader = os.Stdin

	scanner := bufio.NewScanner(reader)

	puzzle := Puzzle{}

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

		puzzle.Board = append(puzzle.Board, cells)
	}

	_, err := validatePuzzle(puzzle)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Puzzle is valid")
	}

	printPuzzle(puzzle)
}

func validatePuzzle(puzzle Puzzle) (bool, error) {
	_, err := checkForInvalidValues(puzzle.Board)
	if err != nil {
		// early exit
		return false, err
	}

	// check each row
	for rowIndex := range gridSize {
		_, err := areaHasDuplicate(puzzle.getRow(rowIndex), Row, rowIndex)
		if err != nil {
			return false, fmt.Errorf("Row check failed: %v", err)
		}
	}

	// check each column
	for columnIndex := range gridSize {
		_, err := areaHasDuplicate(puzzle.getColumn(columnIndex), Column, columnIndex)
		if err != nil {
			return false, fmt.Errorf("Column check failed: %v", err)
		}
	}

	// check each 3x3 sector
	for sectorIndex := range gridSize {
		_, err := areaHasDuplicate(puzzle.getSector(sectorIndex), Sector, sectorIndex)
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

func printPuzzle(puzzle Puzzle) {
	header :=
		"╔═══════╤═══════╤═══════╗"
	sectorDivider :=
		"╠═══════╪═══════╪═══════╣"
	footer :=
		"╚═══════╧═══════╧═══════╝"

	fmt.Println(header)
	for i, row := range puzzle.Board {
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
