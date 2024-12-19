package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// +-----------------+-----------------+-----------------+
// | 0,0   0,1   0,2 | 0,3   0,4   0,5 | 0,6   0,7   0,8 |
// | 1,0   1,1   1,2 | 1,3   1,4   1,5 | 1,6   1,7   1,8 |
// | 2,0   2,1   2,2 | 2,3   2,4   2,5 | 2,6   2,7   2,8 |
// +-----------------+-----------------+-----------------+
// | 3,0   3,1   3,2 | 3,3   3,4   3,5 | 3,6   3,7   3,8 |
// | 4,0   4,1   4,2 | 4,3   4,4   4,5 | 4,6   4,7   4,8 |
// | 5,0   5,1   5,2 | 5,3   5,4   5,5 | 5,6   5,7   5,8 |
// +-----------------+-----------------+-----------------+
// | 6,0   6,1   6,2 | 6,3   6,4   6,5 | 6,6   6,7   6,8 |
// | 7,0   7,1   7,2 | 7,3   7,4   7,5 | 7,6   7,7   7,8 |
// | 8,0   8,1   8,2 | 8,3   8,4   8,5 | 8,6   8,7   8,8 |
// +-----------------+-----------------+-----------------+

func TestValidatePuzzle(t *testing.T) {
	tests := []struct {
		name                  string
		filename              string
		expectedValid         bool
		expectedErrorContains string
	}{
		{
			name:                  "valid puzzle",
			filename:              "samples/001.txt",
			expectedValid:         true,
			expectedErrorContains: "",
		},
		{
			name:                  "invalid row",
			filename:              "samples/invalid_row.txt",
			expectedValid:         false,
			expectedErrorContains: "Row check failed",
		},
		{
			name:                  "invalid column",
			filename:              "samples/invalid_column.txt",
			expectedValid:         false,
			expectedErrorContains: "Column check failed",
		},
		{
			name:                  "invalid sector",
			filename:              "samples/invalid_sector.txt",
			expectedValid:         false,
			expectedErrorContains: "Sector check failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, err := os.ReadFile(tt.filename)
			if err != nil {
				panic(fmt.Sprintf("Unable to read file %s", tt.filename))
			}

			puzzle := hydratePuzzle(string(contents))
			valid, err := validatePuzzle(puzzle)

			assert.Equal(t, valid, tt.expectedValid)
			if !valid {
				assert.Contains(t, err.Error(), tt.expectedErrorContains)
			}
		})
	}
}
