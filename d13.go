package main

import (
	"strings"
)

// validKnightsD13 checks for a valid solution of knight's tour.
func validKnightsD13(solution string) bool {
	const numSteps = 64 // Must be 64 for a 8x8 table.
	sl := strings.Split(solution, "\n")
	var prevStep string = sl[0]
	var row, col int

	// A valid solution must have 64 steps.
	if len(sl) != numSteps {
		return false
	}
	// A valid solution must not have duplicated entries.
	for i, step := range sl {
		for j, testStep := range sl {
			if step == testStep && i != j {
				return false
			}
		}
	}
	// Test limits of table.
	for _, step := range sl {
		if !(step[0] >= 'a' && step[0] <= 'h' && step[1] >= '1' && step[1] <= '8') {
			return false
		}
	}
	// Verify if the move is valid from previous one.
	// Valid moves are:
	//	 1, 2
	//	 2, 1
	//	 1,-2
	//	 2,-1
	//	-1, 2
	//	-2, 1
	//	-1,-2
	//	-2,-1
	// The sum of abs values of the valid moves are always 3.
	for _, step := range sl[1:] {
		col = int(step[0]) - int(prevStep[0])
		if col < 0 {
			col = 0 - col
		}
		row = int(step[1]) - int(prevStep[1])
		if row < 0 {
			row = 0 - row
		}
		// Prevent invalid moves which sum results 3.
		if col < 1 || col > 2 {
			return false
		}
		if col+row != 3 {
			return false
		}
		prevStep = step
	}
	return true
}
