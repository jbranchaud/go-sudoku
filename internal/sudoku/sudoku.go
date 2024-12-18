package sudoku

import (
	"fmt"
	"strconv"
	"strings"
)

// TODO: update all the `Get*` methods to be things like `puzzle.CurrentBoard()`
// and `puzzle.RowAt(rowIndex int)`

var GridSize = 9

type Placement struct {
	Row   int
	Cell  int
	Value int
}

type Puzzle struct {
	Board    [][]int
	Solution []Placement
}

func (puz *Puzzle) String() string {
	board := puz.CurrentBoard()

	var builder strings.Builder
	for i, row := range board {
		for _, cell := range row {
			builder.WriteString(strconv.Itoa(cell))
		}

		// add a new line after each row of cells, unless it is the last row
		if i != len(board)-1 {
			builder.WriteString("\n")
		}
	}

	return builder.String()
}

// Make a copy of the initial puzzle board and apply all Placements in the
// Solution to it.
func (puz *Puzzle) CurrentBoard() [][]int {
	currentBoard := make([][]int, GridSize)
	for i := range GridSize {
		currentBoard[i] = make([]int, GridSize)
		copy(currentBoard[i], puz.Board[i])
	}

	for _, p := range puz.Solution {
		currentBoard[p.Row][p.Cell] = p.Value
	}

	return currentBoard
}

func (puz *Puzzle) GetRow(rowIndex int) []int {
	if rowIndex < 0 || rowIndex > 8 {
		panic(fmt.Sprintf("Invalid rowIndex %d", rowIndex))
	}

	return puz.CurrentBoard()[rowIndex]
}

func (puz *Puzzle) GetColumn(colIndex int) []int {
	column := []int{}
	for _, row := range puz.CurrentBoard() {
		column = append(column, row[colIndex])
	}

	return column
}

func (puz *Puzzle) GetSector(secIndex int) []int {
	sector := []int{}
	for i := range 3 {
		for j := range 3 {
			rowIndex := ((secIndex / 3) * 3) + i
			cellIndex := ((secIndex % 3) * 3) + j

			sector = append(sector, puz.CurrentBoard()[rowIndex][cellIndex])
		}
	}

	return sector
}

func (puz *Puzzle) PlaceValue(row int, cell int, value int) {
	potentialPlacement := Placement{Row: row, Cell: cell, Value: value}
	puz.Solution = append(puz.Solution, potentialPlacement)
}

func (puz *Puzzle) UndoLastPlacement() {
	Pop(puz.Solution)
}
