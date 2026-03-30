package main

import (
	"math/rand"
)

// generateBoard creates a new shuffled board.
func generateBoard(innerRows, innerCols, numKinds, tilesPerKind int, rng *rand.Rand) *Board {
	total := innerRows * innerCols
	if numKinds*tilesPerKind != total {
		panic("tile count mismatch: numKinds * tilesPerKind must equal innerRows * innerCols")
	}

	// Create tile pool
	pool := make([]TileKind, 0, total)
	for k := 1; k <= numKinds; k++ {
		for range tilesPerKind {
			pool = append(pool, TileKind(k))
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

// shuffleRemaining shuffles all remaining tiles in place.
func shuffleRemaining(b *Board, rng *rand.Rand) {
	var positions []Point
	var tiles []TileKind

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
