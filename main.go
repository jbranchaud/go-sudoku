package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var gridSize = 9

type Placement struct {
	Row   int
	Cell  int
	Value int
}

type Puzzle struct {
	Board    [][]int
	Solution []Placement
}

// Make a copy of the initial puzzle board and apply all Placements in the
// Solution to it.
func (puz *Puzzle) getCurrentBoard() [][]int {
	currentBoard := make([][]int, gridSize)
	for i := range gridSize {
		currentBoard[i] = make([]int, gridSize)
		copy(currentBoard[i], puz.Board[i])
	}

	for _, p := range puz.Solution {
		currentBoard[p.Row][p.Cell] = p.Value
	}

	return currentBoard
}

func (puz *Puzzle) getRow(rowIndex int) []int {
	if rowIndex < 0 || rowIndex > 8 {
		panic(fmt.Sprintf("Invalid rowIndex %d", rowIndex))
	}

	return puz.getCurrentBoard()[rowIndex]
}

func (puz *Puzzle) getColumn(colIndex int) []int {
	column := []int{}
	for _, row := range puz.getCurrentBoard() {
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

			sector = append(sector, puz.getCurrentBoard()[rowIndex][cellIndex])
		}
	}

	return sector
}

func readInPuzzle(scanner *bufio.Scanner) Puzzle {
	var puzzle Puzzle
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

	return puzzle
}

type Options struct {
	Debug      bool
	SolveOrder Order
	Seed       int64
	Rng        *rand.Rand
}

func NewOptions(debug bool, solveOrder Order, seedFromFlag *int64) Options {
	options := Options{
		Debug:      debug,
		SolveOrder: solveOrder,
	}

	if solveOrder == Shuffled {
		var seed int64

		if seedFromFlag == nil {
			seed = rand.Int63()
		} else {
			seed = *seedFromFlag
		}

		options.Seed = seed
		options.Rng = rand.New(rand.NewSource(seed))
	}

	return options
}

func main() {
	cmdGenerate := &cobra.Command{
		Use:   "generate",
		Short: "Generate a random, solved puzzle",
		Long:  `Generate a randomly-seeded puzzle that is fully solved`,
		Run: func(cmd *cobra.Command, args []string) {
			var seedFromFlag *int64
			if cmd.Flags().Changed("seed") {
				seed, err := cmd.Flags().GetInt64("seed")
				if err != nil {
					fmt.Println("Seed flag is missing from `cmdFlags()`")
					os.Exit(1)
				}

				seedFromFlag = &seed
			}

			options := NewOptions(false, Shuffled, seedFromFlag)

			puzzle := generateSolvedPuzzle(options)

			fmt.Printf("Generated puzzle with seed %d\n", options.Seed)
			printPuzzle(puzzle)
		},
	}
	cmdSolve := &cobra.Command{
		Use:   "solve [puzzle file]",
		Short: "Solve the given Sudoku puzzle",
		Long:  `A sudoku puzzle given to stdin will be validated and solved`,
		Run: func(cmd *cobra.Command, args []string) {
			debug, err := cmd.Flags().GetBool("debug")
			if err != nil {
				fmt.Println("Debug flag is missing from `cmdFlags()`")
				os.Exit(1)
			}

			var reader io.Reader
			if len(args) > 0 {
				// read the puzzle from the given file
				file, err := os.Open(args[0])
				if err != nil {
					fmt.Printf("Unable to read file: %s\n", args[0])
					os.Exit(1)
				}

				reader = file
			} else {
				file, err := os.Stdin.Stat()
				if err != nil {
					fmt.Printf("Error checking stdin: %v\n", err)
					os.Exit(1)
				}
				waitingForUserInput := (file.Mode() & os.ModeCharDevice) != 0

				if waitingForUserInput {
					fmt.Print("Enter a file name for puzzle to solve: ")
					termInputScanner := bufio.NewScanner(os.Stdin)
					var filename string
					for termInputScanner.Scan() {
						filename = termInputScanner.Text()
						break
					}

					file, err := os.Open(filename)
					if err != nil {
						fmt.Printf("Unable to read file: %s\n", args[0])
						os.Exit(1)
					}

					reader = file
				} else {
					// input is being piped in from a file to stdin
					reader = os.Stdin
				}
			}

			scanner := bufio.NewScanner(reader)
			puzzle := readInPuzzle(scanner)

			options := NewOptions(debug, InOrder, nil)
			solvePuzzle(puzzle, options)
		},
	}
	var Debug bool
	var Seed int64
	var rootCmd = &cobra.Command{Use: "go-sudoku"}
	rootCmd.AddCommand(cmdSolve)
	rootCmd.AddCommand(cmdGenerate)
	cmdGenerate.PersistentFlags().Int64VarP(&Seed, "seed", "", -1, "deterministically seed generated puzzle")
	rootCmd.PersistentFlags().BoolVarP(&Debug, "debug", "", false, "turns on debug mode, extra logging")
	rootCmd.Execute()
}

func generateSolvedPuzzle(options Options) Puzzle {
	board := make([][]int, gridSize)
	for i := range gridSize {
		board[i] = make([]int, gridSize)
	}
	emptyPuzzle := Puzzle{Board: board}

	status, puzzle, _ := traversePuzzle(emptyPuzzle, 1, options, &Diagnostics{})

	if status != Solved {
		fmt.Println("Something went wrong with puzzle generation")
		os.Exit(1)
	}

	return puzzle
}

func solvePuzzle(puzzle Puzzle, options Options) {
	fmt.Println("Initial puzzle:")
	printPuzzle(puzzle)

	_, err := validatePuzzle(puzzle)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Puzzle is valid")
	}

	status, puzzle, diagnostics := traversePuzzle(puzzle, 1, options, &Diagnostics{})
	if status == Solved {
		fmt.Println("Solved the puzzle:")
		printPuzzle(puzzle)
	} else {
		fmt.Println("Unable to solve puzzle:")
		printPuzzle(puzzle)
	}

	if options.Debug {
		fmt.Println("")
		fmt.Println("Search Space Diagnostics:")
		fmt.Printf("Nodes Visited: %d\n", diagnostics.NodeVisitCount)
		fmt.Printf("Backtracks: %d\n", diagnostics.BacktrackCount)
		fmt.Printf("Validity Checks: %d\n", diagnostics.ValidityCheckCount)
	}
}

func validatePuzzle(puzzle Puzzle) (bool, error) {
	_, err := checkForInvalidValues(puzzle.getCurrentBoard())
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

type PuzzleStatus string

const (
	Invalid PuzzleStatus = "Invalid"
	Valid   PuzzleStatus = "Valid"
	Solved  PuzzleStatus = "Solved"
)

type Diagnostics struct {
	BacktrackCount     int
	NodeVisitCount     int
	ValidityCheckCount int
}

func traversePuzzle(puzzle Puzzle, level int, options Options, diagnostics *Diagnostics) (PuzzleStatus, Puzzle, Diagnostics) {
	// this is a recursive function, so:
	// initial pass => puzzle should be Valid
	// cell is filled in =>
	// - if the value makes the puzzle invalid, Invalid
	// - if the value is a valid value, Valid
	// final cell is filled in =>
	// - if the value makes the puzzle invalid, Invalid
	// - if the value solves the puzzle, Solved
	status := checkPuzzleStatus(puzzle)
	(*diagnostics).ValidityCheckCount++

	// max depth of the traversal is the number of cells on the board
	// don't let the traversal exceed it
	maxDepth := gridSize*gridSize + 1
	if level > maxDepth {
		panic(fmt.Sprintf("traversePuzzle:level has exceeded %d", maxDepth))
	}

	switch status {
	case Solved:
		return Solved, puzzle, *diagnostics
	case Valid:
		nextRow, nextCell, err := findNextEmptyCell(puzzle)
		if err != nil {
			panic(fmt.Sprintf("Shouldn't reach here for valid puzzle: %v", err))
		}

		possibleValues := findPossibleValues(puzzle, nextRow, nextCell, options)

		// make another puzzle placement
		for _, value := range possibleValues {
			(*diagnostics).NodeVisitCount++
			potentialPlacement := Placement{Row: nextRow, Cell: nextCell, Value: value}
			puzzle.Solution = append(puzzle.Solution, potentialPlacement)

			if options.Debug {
				fmt.Printf("%d) placing %d at (%d,%d) of %v\n", level, value, nextRow, nextCell, possibleValues)
			}

			latestStatus, latestPuzzle, _ := traversePuzzle(puzzle, level+1, options, diagnostics)
			switch latestStatus {
			case Solved:
				return Solved, latestPuzzle, *diagnostics
			case Invalid:
				// undo latest placement, continue
				(*diagnostics).BacktrackCount++
				Pop(latestPuzzle.Solution)
				continue
			default:
				// we shouldn't get here, something went wrong
				panic("traversePuzzle returned an unrecognized status")
			}
		}

		// if we haven't found a solution at this point, then we'll need to backtrack
		return Invalid, puzzle, *diagnostics
	case Invalid:
		return Invalid, puzzle, *diagnostics
	default:
		panic("Should not have reached here when traversing puzzle")
	}
}

func checkPuzzleStatus(puzzle Puzzle) PuzzleStatus {
	valid, err := validatePuzzle(puzzle)
	if err != nil {
		return Invalid
	}

	if !valid {
		return Invalid
	} else {
		_, _, err := findNextEmptyCell(puzzle)
		if err != nil {
			// we are valid, but there are no more empty cells
			return Solved
		} else {
			return Valid
		}
	}
}

func findNextEmptyCell(puzzle Puzzle) (int, int, error) {
	currentBoard := puzzle.getCurrentBoard()

	for row := range gridSize {
		for cell := range gridSize {
			if currentBoard[row][cell] == 0 {
				return row, cell, nil
			}
		}
	}

	return -1, -1, fmt.Errorf("No more empty cells in the puzzle")
}

type Order string

const (
	InOrder  Order = "InOrder"
	Shuffled Order = "Shuffled"
)

func findPossibleValues(puzzle Puzzle, row int, cell int, options Options) []int {
	usedValues := make(map[int]int)

	sectorNum := GetSectorNumberForCell(row, cell)

	cellsConstrainingThisCell := slices.Concat(
		puzzle.getRow(row),
		puzzle.getColumn(cell),
		puzzle.getSector(sectorNum),
	)
	for _, rowEntry := range cellsConstrainingThisCell {
		if rowEntry != 0 {
			usedValues[rowEntry]++
		}
	}

	unusedValues := []int{}

	for i := range gridSize {
		value := i + 1
		if usedValues[value] == 0 {
			unusedValues = append(unusedValues, value)
		}
	}

	if options.SolveOrder == Shuffled {
		Shuffle(unusedValues, options.Rng)
	}

	return unusedValues
}

func listMissingValues(section []int) []int {
	missingValues := []int{}

	for i := range gridSize {
		val := i + 1
		seen := false
		for _, cell := range section {
			if cell == val {
				seen = true
				break
			}
		}

		if !seen {
			missingValues = append(missingValues, val)
		}
	}

	return missingValues
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

	currentBoard := puzzle.getCurrentBoard()

	fmt.Println(header)
	for i, row := range currentBoard {
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
