package soduku

import "github.com/laurentlp/sudoku-solver/solver"

func Solve(grid string) {
	if _, err := solver.Solve(grid); err != nil {
		panic(err)
	}
}
