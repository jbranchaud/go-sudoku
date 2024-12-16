# go-sudoku

> Sudoku from the CLI with Go

## Representing a puzzle

Create a file that is made up of nine lines of nine characters each. 1 through
9 for filled in values. 0 for blank cells.

E.g.

```bash
$ cat samples/001.txt
000080000
823107496
000000008
948002001
075000600
601049820
080010902
000763000
510928074
```

## Solve

The `solve` command will read in a puzzle and solve it.

You can invoke it one of the following ways:
- specify the puzzle file for it to read in
- redirect a puzzle to it via stdin
- no arg, it will wait for a puzzle file name to be entered

```bash
$ go run . solve samples/001.txt
$ go run . solve < samples/001.txt
$ go run . solve
Enter a file name for puzzle to solve: samples/001.txt

Initial puzzle:
╔═══════╤═══════╤═══════╗
║ _ _ _ │ _ 8 _ │ _ _ _ ║
║ 8 2 3 │ 1 _ 7 │ 4 9 6 ║
║ _ _ _ │ _ _ _ │ _ _ 8 ║
╠═══════╪═══════╪═══════╣
║ 9 4 8 │ _ _ 2 │ _ _ 1 ║
║ _ 7 5 │ _ _ _ │ 6 _ _ ║
║ 6 _ 1 │ _ 4 9 │ 8 2 _ ║
╠═══════╪═══════╪═══════╣
║ _ 8 _ │ _ 1 _ │ 9 _ 2 ║
║ _ _ _ │ 7 6 3 │ _ _ _ ║
║ 5 1 _ │ 9 2 8 │ _ 7 4 ║
╚═══════╧═══════╧═══════╝
Puzzle is valid
Solved the puzzle:
╔═══════╤═══════╤═══════╗
║ 1 6 9 │ 2 8 4 │ 7 5 3 ║
║ 8 2 3 │ 1 5 7 │ 4 9 6 ║
║ 7 5 4 │ 3 9 6 │ 2 1 8 ║
╠═══════╪═══════╪═══════╣
║ 9 4 8 │ 6 7 2 │ 5 3 1 ║
║ 2 7 5 │ 8 3 1 │ 6 4 9 ║
║ 6 3 1 │ 5 4 9 │ 8 2 7 ║
╠═══════╪═══════╪═══════╣
║ 3 8 7 │ 4 1 5 │ 9 6 2 ║
║ 4 9 2 │ 7 6 3 │ 1 8 5 ║
║ 5 1 6 │ 9 2 8 │ 3 7 4 ║
╚═══════╧═══════╧═══════╝
```
