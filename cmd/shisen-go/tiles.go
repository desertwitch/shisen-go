package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// initTileCache builds a cache of pre-rendered tiles for repeat usage.
func initTileCache(face text.Face) {
	tileImageCache = make(map[TileSymbol]*ebiten.Image)
	tileSelectedCache = make(map[TileSymbol]*ebiten.Image)

	for k := TileSymbol(1); k <= TileSymbol(gameNumSymbols); k++ {
		tileImageCache[k] = renderTile(k, false, face)
		tileSelectedCache[k] = renderTile(k, true, face)
	}
}

// renderTile renders one tile as an [ebiten.Image] to be drawn later.
func renderTile(sym TileSymbol, selected bool, face text.Face) *ebiten.Image {
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
	vector.StrokeRect(img, 0, 0, float32(tileW-2), float32(tileH-2), tileBorderW, borderColor, false)

	// Symbol
	idx := int(sym) - 1
	if idx < 0 || idx >= len(tileSymbols) {
		idx = 0
	}
	symText := tileSymbols[idx]
	symColor := tileColors[(idx/len(tileColors))%len(tileColors)]

	op := &text.DrawOptions{}
	w, h := text.Measure(symText, face, 0)
	op.GeoM.Translate(
		float64(tileW-2)/2-w/2,
		float64(tileH-2)/2-h/2,
	)
	op.ColorScale.ScaleWithColor(symColor)
	text.Draw(img, symText, face, op)

	return img
}

// drawTile draws a pre-rendered tile at given pixel coordinates.
func drawTile(screen *ebiten.Image, sym TileSymbol, x, y float64, selected bool) {
	if sym == TileEmpty {
		return
	}

	var img *ebiten.Image

	if selected {
		img = tileSelectedCache[sym]
	} else {
		img = tileImageCache[sym]
	}

	if img == nil {
		return
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)
	screen.DrawImage(img, op)
}

// drawPath draws the connecting path as a yellow line on the screen.
func drawPath(screen *ebiten.Image, path []Point, offsetX, offsetY float64) {
	if len(path) < 2 {
		return
	}

	for i := range len(path) - 1 {
		x1 := offsetX + float64(path[i].C)*tileW + tileW/2
		y1 := offsetY + float64(path[i].R)*tileH + tileH/2
		x2 := offsetX + float64(path[i+1].C)*tileW + tileW/2
		y2 := offsetY + float64(path[i+1].R)*tileH + tileH/2

		vector.StrokeLine(screen, float32(x1), float32(y1), float32(x2), float32(y2), pathW, pathColor, false)
	}
}
