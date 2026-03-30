package main

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var suitColors = []color.RGBA{
	{0xC8, 0x50, 0x50, 0xFF}, // red
	{0x50, 0x90, 0xC8, 0xFF}, // blue
	{0x50, 0x98, 0x60, 0xFF}, // green
	{0x98, 0x78, 0x30, 0xFF}, // gold
	{0x80, 0x58, 0xB0, 0xFF}, // purple
	{0xC8, 0x78, 0x40, 0xFF}, // orange
}

var tileSymbols = []string{
	"\u0410", "\u0414", "\u0416", "\u0428", "\u042E", "\u042F", // А Д Ж Ш Ю Я
	"\u0411", "\u0417", "\u041B", "\u041F", "\u0424", "\u0427", // Б З Л П Ф Ч
	"\u0412", "\u0413", "\u0418", "\u041C", "\u0420", "\u0421", // В Г И М Р С
	"\u03B1", "\u03B2", "\u03B3", "\u03B4", "\u03B5", "\u03B6", // α β γ δ ε ζ
	"R", "W", "Z", "F", "G", "X",
	"1", "2", "3", "4", "5", "6",
}

var (
	tileImageCache    map[TileKind]*ebiten.Image
	tileSelectedCache map[TileKind]*ebiten.Image
)

func initTileCache(face text.Face) {
	tileImageCache = make(map[TileKind]*ebiten.Image)
	tileSelectedCache = make(map[TileKind]*ebiten.Image)

	for k := TileKind(1); k <= TileKind(NumTileKinds); k++ {
		tileImageCache[k] = renderTile(k, false, face)
		tileSelectedCache[k] = renderTile(k, true, face)
	}
}

func renderTile(kind TileKind, selected bool, face text.Face) *ebiten.Image {
	img := ebiten.NewImage(TileW, TileH)

	// Tile background
	bgColor := color.RGBA{0xF5, 0xF0, 0xE0, 0xFF} // cream
	if selected {
		bgColor = color.RGBA{0xFF, 0xFF, 0x80, 0xFF} // bright yellow highlight
	}

	// Shadow
	shadowColor := color.RGBA{0x80, 0x78, 0x68, 0xFF} // gray
	vector.DrawFilledRect(img, 2, 2, float32(TileW-2), float32(TileH-2), shadowColor, false)

	// Main tile body
	vector.DrawFilledRect(img, 0, 0, float32(TileW-2), float32(TileH-2), bgColor, false)

	// Border
	borderColor := color.RGBA{0x60, 0x58, 0x48, 0xFF} // brown
	if selected {
		borderColor = color.RGBA{0xE0, 0xA0, 0x00, 0xFF} // bright yellow highlight
	}
	vector.StrokeRect(img, 0, 0, float32(TileW-2), float32(TileH-2), 1.5, borderColor, false)

	// Symbol
	idx := int(kind) - 1
	if idx < 0 || idx >= len(tileSymbols) {
		idx = 0
	}
	sym := tileSymbols[idx]
	suitIdx := idx / 6
	symColor := suitColors[suitIdx%len(suitColors)]

	// Draw the symbol centered
	op := &text.DrawOptions{}
	w, h := text.Measure(sym, face, 0)
	op.GeoM.Translate(
		float64(TileW-2)/2-w/2,
		float64(TileH-2)/2-h/2,
	)
	op.ColorScale.ScaleWithColor(symColor)
	text.Draw(img, sym, face, op)

	return img
}

// DrawTile draws a tile at pixel coordinates.
func DrawTile(screen *ebiten.Image, kind TileKind, x, y float64, selected bool) {
	if kind == TileEmpty {
		return
	}

	var img *ebiten.Image

	if selected {
		img = tileSelectedCache[kind]
	} else {
		img = tileImageCache[kind]
	}

	if img == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

// DrawPath draws the connecting path as a yellow line on the screen.
func DrawPath(screen *ebiten.Image, path []Point, offsetX, offsetY float64) {
	if len(path) < 2 {
		return
	}

	pathColor := color.RGBA{0xFF, 0xC0, 0x00, 0xFF} // gold
	for i := 0; i < len(path)-1; i++ {
		x1 := offsetX + float64(path[i].C)*TileW + TileW/2
		y1 := offsetY + float64(path[i].R)*TileH + TileH/2
		x2 := offsetX + float64(path[i+1].C)*TileW + TileW/2
		y2 := offsetY + float64(path[i+1].R)*TileH + TileH/2

		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3, pathColor, false)
	}
}

// CellRect returns the bounding rectangle for a board cell.
func CellRect(r, c int, offsetX, offsetY float64) image.Rectangle {
	x := int(offsetX) + c*TileW
	y := int(offsetY) + r*TileH

	return image.Rect(x, y, x+TileW, y+TileH)
}
