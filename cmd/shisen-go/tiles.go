package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var tileSymbols = []string{
	"\u0410", "\u0414", "\u0416", "\u0428", "\u042E", "\u042F", // А Д Ж Ш Ю Я
	"\u0411", "\u0417", "\u041B", "\u041F", "\u0424", "\u0427", // Б З Л П Ф Ч
	"\u0412", "\u0413", "\u0418", "\u041C", "\u0420", "\u0421", // В Г И М Р С
	"\u03B1", "\u03B2", "\u03B3", "\u03B4", "\u03B5", "\u03B6", // α β γ δ ε ζ
	"R", "W", "Z", "F", "G", "X",
	"1", "2", "4", "5", "6", "7",
}

var tileColors = []color.RGBA{
	{0xC8, 0x50, 0x50, 0xFF}, // red
	{0x50, 0x90, 0xC8, 0xFF}, // blue
	{0x50, 0x98, 0x60, 0xFF}, // green
	{0x98, 0x78, 0x30, 0xFF}, // gold
	{0x80, 0x58, 0xB0, 0xFF}, // purple
	{0xC8, 0x78, 0x40, 0xFF}, // orange
}

var (
	tileImageCache    map[TileKind]*ebiten.Image
	tileSelectedCache map[TileKind]*ebiten.Image

	tileColorBackground         = color.RGBA{0xF5, 0xF0, 0xE0, 0xFF} // cream
	tileColorBackgroundSelected = color.RGBA{0xFF, 0xFF, 0x80, 0xFF} // bright yellow
	tileColorShadow             = color.RGBA{0x80, 0x78, 0x68, 0xFF} // gray
	tileColorBorder             = color.RGBA{0x60, 0x58, 0x48, 0xFF} // brown
	tileColorBorderSelected     = color.RGBA{0xE0, 0xA0, 0x00, 0xFF} // bright yellow

	pathColor = color.RGBA{0xFF, 0xC0, 0x00, 0xFF} // gold
)

func initTileCache(face text.Face) {
	tileImageCache = make(map[TileKind]*ebiten.Image)
	tileSelectedCache = make(map[TileKind]*ebiten.Image)

	for k := TileKind(1); k <= TileKind(numTileKinds); k++ {
		tileImageCache[k] = renderTile(k, false, face)
		tileSelectedCache[k] = renderTile(k, true, face)
	}
}

//nolint:mnd
func renderTile(kind TileKind, selected bool, face text.Face) *ebiten.Image {
	img := ebiten.NewImage(tileW, tileH)

	// Background
	bgColor := tileColorBackground
	if selected {
		bgColor = tileColorBackgroundSelected
	}

	// Shadow
	shadowColor := tileColorShadow
	vector.DrawFilledRect(img, 2, 2, float32(tileW-2), float32(tileH-2), shadowColor, false)

	// Body
	vector.DrawFilledRect(img, 0, 0, float32(tileW-2), float32(tileH-2), bgColor, false)

	// Border
	borderColor := tileColorBorder
	if selected {
		borderColor = tileColorBorderSelected
	}
	vector.StrokeRect(img, 0, 0, float32(tileW-2), float32(tileH-2), 1.5, borderColor, false)

	// Symbol
	idx := int(kind) - 1
	if idx < 0 || idx >= len(tileSymbols) {
		idx = 0
	}
	sym := tileSymbols[idx]
	suitIdx := idx / 6
	symColor := tileColors[suitIdx%len(tileColors)]

	op := &text.DrawOptions{}
	w, h := text.Measure(sym, face, 0)
	op.GeoM.Translate(
		float64(tileW-2)/2-w/2,
		float64(tileH-2)/2-h/2,
	)
	op.ColorScale.ScaleWithColor(symColor)
	text.Draw(img, sym, face, op)

	return img
}

// drawTile draws a tile at pixel coordinates.
func drawTile(screen *ebiten.Image, kind TileKind, x, y float64, selected bool) {
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

// drawPath draws the connecting path as a yellow line on the screen.
//
//nolint:mnd
func drawPath(screen *ebiten.Image, path []Point, offsetX, offsetY float64) {
	if len(path) < 2 {
		return
	}

	for i := range len(path) - 1 {
		x1 := offsetX + float64(path[i].C)*tileW + tileW/2
		y1 := offsetY + float64(path[i].R)*tileH + tileH/2
		x2 := offsetX + float64(path[i+1].C)*tileW + tileW/2
		y2 := offsetY + float64(path[i+1].R)*tileH + tileH/2

		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), 3, pathColor, false)
	}
}
