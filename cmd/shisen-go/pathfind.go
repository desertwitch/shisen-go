package main

// Point is a grid coordinate.
type Point struct {
	R, C int
}

// Direction vectors: up, down, left, right.
var dirs = [4]Point{
	{-1, 0}, {1, 0}, {0, -1}, {0, 1},
}

// FindPath checks whether tiles at (r1,c1) and (r2,c2) can be connected.
// It returns the path as corner points (including start and end), or nil.
func FindPath(b *Board, r1, c1, r2, c2 int) []Point {
	// Both selected tiles are actually the same tile:
	if r1 == r2 && c1 == c2 {
		return nil
	}
	// One of the selected tiles is an empty tile:
	if b.Get(r1, c1) == TileEmpty || b.Get(r2, c2) == TileEmpty {
		return nil
	}
	// Both selected tiles are not of the same kind:
	if b.Get(r1, c1) != b.Get(r2, c2) {
		return nil
	}

	start := Point{r1, c1}
	end := Point{r2, c2}

	// 0 bends (vertical + horizontal):
	//
	// {R1.C1} ------------ {R2.C2}
	//
	if clearBetween(b, r1, c1, r2, c2) {
		return []Point{start, end}
	}

	// 1 bends (vertical):
	//
	// {R1.C1}
	//    |
	// {R2.C1} ------------ {R2.C2}
	//
	if b.IsEmpty(r2, c1) && clearBetween(b, r1, c1, r2, c1) && clearBetween(b, r2, c1, r2, c2) {
		return []Point{start, {r2, c1}, end}
	}
	// 1 bends (horizontal):
	//
	// {R1.C1} ------------ {R1.C2}
	//                         |
	//                      {R2.C2}
	//
	if b.IsEmpty(r1, c2) && clearBetween(b, r1, c1, r1, c2) && clearBetween(b, r1, c2, r2, c2) {
		return []Point{start, {r1, c2}, end}
	}

	// Walk a ray from the start in each direction, for every empty {MR.MC} on
	// the way we then try to find a 1-bend connection towards the actual end point.
	for _, d := range dirs {
		mr, mc := r1+d.R, c1+d.C
		for mr >= 0 && mr < b.Rows && mc >= 0 && mc < b.Cols && b.IsEmpty(mr, mc) {
			// 2 bends (vertical):
			//
			// {R1.C1}
			//    |
			// {MR.C1} ------------ {MR.C2}
			//                         |
			//                      {R2.C2}
			//
			if d.C == 0 && b.IsEmpty(mr, c2) &&
				clearBetween(b, mr, c1, mr, c2) && clearBetween(b, mr, c2, r2, c2) {
				return []Point{start, {mr, c1}, {mr, c2}, end}
			}
			// 2 bends (horizontal):
			//
			// {R1.C1} ------------ {R1.MC}
			//                         |
			//                      {R2.MC} ------------ {R2.C2}
			//
			if d.R == 0 && b.IsEmpty(r2, mc) &&
				clearBetween(b, r1, mc, r2, mc) && clearBetween(b, r2, mc, r2, c2) {
				return []Point{start, {r1, mc}, {r2, mc}, end}
			}

			mr += d.R
			mc += d.C
		}
	}

	return nil
}

// clearBetween checks if all cells between (r1,c1) and (r2,c2) are empty.
// The endpoints are excluded; checks for axis-aligned lines (no diagonals).
func clearBetween(b *Board, r1, c1, r2, c2 int) bool {
	// Horizontal connection
	if r1 == r2 {
		lo, hi := min(c1, c2), max(c1, c2)
		for c := lo + 1; c < hi; c++ {
			if !b.IsEmpty(r1, c) {
				return false
			}
		}
		return true
	}

	// Vertical connection
	if c1 == c2 {
		lo, hi := min(r1, r2), max(r1, r2)
		for r := lo + 1; r < hi; r++ {
			if !b.IsEmpty(r, c1) {
				return false
			}
		}
		return true
	}

	return false
}

// HasAnyMatch checks whether any valid pair exists on the board.
func HasAnyMatch(b *Board) (bool, Point, Point) {
	r0, c0, r1, c1 := b.InnerBounds()

	// Group tile positions by kind
	groups := make(map[TileKind][]Point)
	for r := r0; r < r1; r++ {
		for c := c0; c < c1; c++ {
			k := b.Cells[r][c]
			if k != TileEmpty {
				groups[k] = append(groups[k], Point{r, c})
			}
		}
	}

	// Check all possible combinations
	for _, positions := range groups {
		for i := 0; i < len(positions); i++ {
			for j := i + 1; j < len(positions); j++ {
				src, dst := positions[i], positions[j]
				if FindPath(b, src.R, src.C, dst.R, dst.C) != nil {
					return true, src, dst
				}
			}
		}
	}

	return false, Point{}, Point{}
}
