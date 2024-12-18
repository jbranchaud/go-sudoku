package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"io"
	"math/rand"
	"os"
	"slices"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/cobra"

	"github.com/jbranchaud/go-sudoku/internal/sudoku"
)

func readInPuzzle(scanner *bufio.Scanner) sudoku.Puzzle {
	var puzzle sudoku.Puzzle
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

func setupDatabase() *sql.DB {
	databaseString := os.Getenv("GOOSE_DBSTRING")
	if len(databaseString) == 0 {
		fmt.Println("Error retrieving `GOOSE_DBSTRING` from env")
		os.Exit(1)
	}
	db, err := sql.Open("sqlite3", databaseString)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		os.Exit(1)
	}

	return db
}

func recordPuzzleTemplate(db *sql.DB, puzzle sudoku.Puzzle, seed int64) int64 {
	insertPuzzleTemplate := `insert into puzzle_templates (seed, board)
		values (?, ?);`

	result, err := db.Exec(insertPuzzleTemplate, seed, puzzle.String())
	if err != nil {
		fmt.Printf("Error inserting puzzle template: %v\n", err)
		os.Exit(1)
	}

	id, _ := result.LastInsertId()

	return id
}

func main() {
	cmdGenerate := &cobra.Command{
		Use:   "generate",
		Short: "Generate a random, solved puzzle",
		Long:  `Generate a randomly-seeded puzzle that is fully solved`,
		Run: func(cmd *cobra.Command, args []string) {
			db := setupDatabase()
			defer db.Close()

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

			id := recordPuzzleTemplate(db, puzzle, options.Seed)

			fmt.Printf("Generated puzzle with seed %d\n", options.Seed)
			printPuzzle(puzzle)
			fmt.Printf("Inserted row in puzzle_templates, id: %d\n", id)
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

func generateSolvedPuzzle(options Options) sudoku.Puzzle {
	board := make([][]int, sudoku.GridSize)
	for i := range sudoku.GridSize {
		board[i] = make([]int, sudoku.GridSize)
	}
	emptyPuzzle := sudoku.Puzzle{Board: board}

	status, puzzle, _ := traversePuzzle(emptyPuzzle, 1, options, &Diagnostics{})

	if status != Solved {
		fmt.Println("Something went wrong with puzzle generation")
		os.Exit(1)
	}

	return puzzle
}

func solvePuzzle(puzzle sudoku.Puzzle, options Options) {
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

func validatePuzzle(puzzle sudoku.Puzzle) (bool, error) {
	_, err := checkForInvalidValues(puzzle.CurrentBoard())
	if err != nil {
		// early exit
		return false, err
	}

	// check each row
	for rowIndex := range sudoku.GridSize {
		_, err := areaHasDuplicate(puzzle.RowAt(rowIndex), Row, rowIndex)
		if err != nil {
			return false, fmt.Errorf("Row check failed: %v", err)
		}
	}

	// check each column
	for columnIndex := range sudoku.GridSize {
		_, err := areaHasDuplicate(puzzle.ColumnAt(columnIndex), Column, columnIndex)
		if err != nil {
			return false, fmt.Errorf("Column check failed: %v", err)
		}
	}

	// check each 3x3 sector
	for sectorIndex := range sudoku.GridSize {
		_, err := areaHasDuplicate(puzzle.SectorAt(sectorIndex), Sector, sectorIndex)
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

func traversePuzzle(puzzle sudoku.Puzzle, level int, options Options, diagnostics *Diagnostics) (PuzzleStatus, sudoku.Puzzle, Diagnostics) {
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
	maxDepth := sudoku.GridSize*sudoku.GridSize + 1
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
			puzzle.PlaceValue(nextRow, nextCell, value)

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
				latestPuzzle.UndoLastPlacement()
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

func checkPuzzleStatus(puzzle sudoku.Puzzle) PuzzleStatus {
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

func findNextEmptyCell(puzzle sudoku.Puzzle) (int, int, error) {
	currentBoard := puzzle.CurrentBoard()

	for row := range sudoku.GridSize {
		for cell := range sudoku.GridSize {
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

func findPossibleValues(puzzle sudoku.Puzzle, row int, cell int, options Options) []int {
	usedValues := make(map[int]int)

	sectorNum := GetSectorNumberForCell(row, cell)

	cellsConstrainingThisCell := slices.Concat(
		puzzle.RowAt(row),
		puzzle.ColumnAt(cell),
		puzzle.SectorAt(sectorNum),
	)
	for _, rowEntry := range cellsConstrainingThisCell {
		if rowEntry != 0 {
			usedValues[rowEntry]++
		}
	}

	unusedValues := []int{}

	for i := range sudoku.GridSize {
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

	for i := range sudoku.GridSize {
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

func printPuzzle(puzzle sudoku.Puzzle) {
	header :=
		"╔═══════╤═══════╤═══════╗"
	sectorDivider :=
		"╠═══════╪═══════╪═══════╣"
	footer :=
		"╚═══════╧═══════╧═══════╝"

	currentBoard := puzzle.CurrentBoard()

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
