package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// initTileCache builds a cache of pre-rendered tiles for repeat usage.
func initTileCache(face text.Face) {
	tileImageCache = make(map[TileKind]*ebiten.Image)
	tileSelectedCache = make(map[TileKind]*ebiten.Image)

	for k := TileKind(1); k <= TileKind(numTileKinds); k++ {
		tileImageCache[k] = renderTile(k, false, face)
		tileSelectedCache[k] = renderTile(k, true, face)
	}
}

// renderTile renders one tile as an [ebiten.Image] to be drawn later.
//
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

// drawTile draws a pre-rendered tile at given pixel coordinates.
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
