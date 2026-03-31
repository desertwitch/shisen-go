package main

import (
	"math/rand"
)

// solvePair records one matching move: two positions sharing the same tile symbol.
type solvePair struct {
	A, B   Point
	Symbol TileSymbol
}

// generateSolvableBoard generates a new shuffled board that was solvable at generation time.
func generateSolvableBoard(innerRows, innerCols, numSymbols, tilesPerSymbol int, rng *rand.Rand) *Board {
	for {
		b := generateBoard(innerRows, innerCols, numSymbols, tilesPerSymbol, rng)
		if solveBoard(b) != nil {
			return b
		}
	}
}

// generateBoard creates a new shuffled board.
func generateBoard(innerRows, innerCols, numSymbols, tilesPerSymbol int, rng *rand.Rand) *Board {
	total := innerRows * innerCols
	if numSymbols*tilesPerSymbol != total {
		panic("tile count mismatch: numSymbols * tilesPerSymbol must equal innerRows * innerCols")
	}

	// Create tile pool
	pool := make([]TileSymbol, 0, total)
	for k := 1; k <= numSymbols; k++ {
		for range tilesPerSymbol {
			pool = append(pool, TileSymbol(k))
		}
	}

	// Shuffle the tiles
	rng.Shuffle(len(pool), func(i, j int) {
		pool[i], pool[j] = pool[j], pool[i]
	})

	// Populate the board
	b := NewBoard(innerRows, innerCols)
	idx := 0
	r0, c0, r1, c1 := b.InnerBounds()
	for r := r0; r < r1; r++ {
		for c := c0; c < c1; c++ {
			b.Cells[r][c] = pool[idx]
			idx++
		}
	}

	return b
}

// solveBoard returns a sequence that clears the board, or nil if unsolvable.
// This is a greedy solver: it repeatedly finds any valid pair and removes it.
func solveBoard(b *Board) []solvePair {
	work := NewBoard(b.Rows-2, b.Cols-2)
	r0, _, r1, _ := b.InnerBounds()
	for r := r0; r < r1; r++ {
		copy(work.Cells[r], b.Cells[r])
	}

	var moves []solvePair
	for {
		found, a, b := hasAnyMatch(work)
		if !found {
			break
		}
		sym := work.Cells[a.R][a.C]
		work.Cells[a.R][a.C] = TileEmpty
		work.Cells[b.R][b.C] = TileEmpty
		moves = append(moves, solvePair{A: a, B: b, Symbol: sym})
	}

	if work.RemainingTiles() != 0 {
		return nil // Unsolvable
	}

	return moves
}

// shuffleRemaining shuffles all remaining tiles in place.
func shuffleRemaining(b *Board, rng *rand.Rand) {
	var positions []Point
	var tiles []TileSymbol

	// Pool the remaining tiles
	r0, c0, r1, c1 := b.InnerBounds()
	for r := r0; r < r1; r++ {
		for c := c0; c < c1; c++ {
			if b.Cells[r][c] != TileEmpty {
				positions = append(positions, Point{r, c})
				tiles = append(tiles, b.Cells[r][c])
			}
		}
	}

	// Shuffle the tiles
	rng.Shuffle(len(tiles), func(i, j int) {
		tiles[i], tiles[j] = tiles[j], tiles[i]
	})

	// Repopulate the board
	for i, p := range positions {
		b.Cells[p.R][p.C] = tiles[i]
	}
}
