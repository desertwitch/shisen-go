package main

// TileKind represents a type of mahjong tile.
type TileKind int

const (
	TileEmpty TileKind = 0
)

// Board holds the 2D grid of tiles.
// A 1-cell empty border is present on every side.
type Board struct {
	Rows  int // total rows including border
	Cols  int // total cols including border
	Cells [][]TileKind
}

// NewBoard creates a board with the given inner dimensions (playable area).
// A 1-cell empty border is added on every side.
func NewBoard(innerRows, innerCols int) *Board {
	rows := innerRows + 2
	cols := innerCols + 2

	cells := make([][]TileKind, rows)

	for r := range cells {
		cells[r] = make([]TileKind, cols)
	}

	return &Board{Rows: rows, Cols: cols, Cells: cells}
}

// InnerBounds returns the top-left and bottom-right of the playable area.
func (b *Board) InnerBounds() (r0, c0, r1, c1 int) {
	return 1, 1, b.Rows - 1, b.Cols - 1
}

// Get returns the tile at (r,c) or TileEmpty (out of bounds).
func (b *Board) Get(r, c int) TileKind {
	if r < 0 || r >= b.Rows || c < 0 || c >= b.Cols {
		return TileEmpty
	}

	return b.Cells[r][c]
}

// Set places a tile kind at (r,c).
func (b *Board) Set(r, c int, k TileKind) {
	b.Cells[r][c] = k
}

// IsEmpty checks if the cell at (r,c) is empty.
func (b *Board) IsEmpty(r, c int) bool {
	return b.Get(r, c) == TileEmpty
}

// RemainingTiles counts non-empty tiles on the board.
func (b *Board) RemainingTiles() int {
	count := 0

	r0, c0, r1, c1 := b.InnerBounds()
	for r := r0; r < r1; r++ {
		for c := c0; c < c1; c++ {
			if b.Cells[r][c] != TileEmpty {
				count++
			}
		}
	}

	return count
}
