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

The default functionality is to read in a puzzle and solve it. Redirect a file
with a puzzle in it to the program and it will solve it.

```bash
$ go run . < samples/001.txt
```
