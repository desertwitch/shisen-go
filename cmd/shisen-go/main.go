package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"image/color"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	ScreenW = 960
	ScreenH = 650
)

const (
	HudTopH = 32 // Pixel height of the top HUD
	HudBotH = 32 // Pixel height of the bottom HUD
	HudPadY = 4  // Pixel height of HUD padding
	TileW   = 48 // Pixel width of a tile
	TileH   = 64 // Pixel height of a tile
)

// Must satisfy: NumTileKinds * TilesPerKind == DefaultInnerRows * DefaultInnerCols
const (
	DefaultInnerRows = 8
	DefaultInnerCols = 18
	NumTileKinds     = 36
	TilesPerKind     = 4
)

const (
	StatePlaying = iota
	StateWin
	StateStuck
)

//go:embed music.ogg
var musicOgg []byte

var musicPlayer *audio.Player

var _ ebiten.Game = (*Game)(nil)

var HudColorDefault = color.RGBA{0x60, 0x60, 0x60, 0xFF}
var HudColorHighlight = color.RGBA{0xFF, 0xFF, 0x80, 0xFF}

type Game struct {
	board      *Board
	rng        *rand.Rand
	state      int // StatePlaying, StateWin, StateStuck
	audioMuted bool

	face   text.Face // Font face for tile symbols
	uiFace text.Face // Font face for UI text

	offsetX float64 // Pixel offset X to center board
	offsetY float64 // Pixel offset Y to center board

	sel1    *Point  // First selected tile (nil if none)
	path    []Point // Last matched path (for animation)
	pathTTL int     // Ticks remaining to show the path

	hint1   *Point // Hint tile 1
	hint2   *Point // Hint tile 2
	hintTTL int    // Ticks remaining to show the hints

	shuffles int // Remaining shuffles

	message string // Message (shuffled, stuck, etc...)
	msgTTL  int    // Ticks remaining to show the message
}

func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	board := GenerateBoard(DefaultInnerRows, DefaultInnerCols, NumTileKinds, TilesPerKind, rng)

	g := &Game{
		board:    board,
		rng:      rng,
		state:    StatePlaying,
		shuffles: 3,
	}

	// Calculate offsets to center the board
	boardPixelW := float64(board.Cols) * TileW
	boardPixelH := float64(board.Rows) * TileH

	g.offsetX = (ScreenW - boardPixelW) / 2
	g.offsetY = HudTopH + (ScreenH-HudTopH-HudBotH-boardPixelH)/2

	return g
}

func (g *Game) initFonts() {
	src, err := text.NewGoTextFaceSource(bytes.NewReader(goRegularTTF))
	if err != nil {
		log.Fatal(err)
	}

	g.face = &text.GoTextFace{
		Source: src,
		Size:   28,
	}

	g.uiFace = &text.GoTextFace{
		Source: src,
		Size:   18,
	}
}

func (g *Game) Layout(_, _ int) (int, int) {
	return ScreenW, ScreenH
}

func (g *Game) Update() error {
	if g.pathTTL > 0 {
		g.pathTTL--
		if g.pathTTL == 0 {
			g.path = nil
		}
	}
	if g.hintTTL > 0 {
		g.hintTTL--
		if g.hintTTL == 0 {
			g.hint1 = nil
			g.hint2 = nil
		}
	}
	if g.msgTTL > 0 {
		g.msgTTL--
		if g.msgTTL == 0 {
			g.message = ""
		}
	}

	if g.state == StatePlaying {
		if inpututil.IsKeyJustPressed(ebiten.KeyH) { // Hint
			g.doHint()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyS) { // Shuffle
			g.doShuffle()
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) { // Mouse
			mx, my := ebiten.CursorPosition()
			g.handleClick(mx, my)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyA) { // Audio
		g.audioMuted = !g.audioMuted
		if g.audioMuted {
			musicPlayer.Pause()
		} else {
			musicPlayer.Play()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyR) { // Restart
		*g = *NewGame()
		g.initFonts()
		initTileCache(g.face)
		return nil
	}

	return nil
}

func (g *Game) handleClick(mx, my int) {
	// Convert pixel to board coordinates
	col := int(float64(mx)-g.offsetX) / TileW
	row := int(float64(my)-g.offsetY) / TileH

	r0, c0, r1, c1 := g.board.InnerBounds()
	if col < 0 || row < 0 || row < r0 || row >= r1 || col < c0 || col >= c1 {
		return // Outside playable area...
	}
	if g.board.IsEmpty(row, col) {
		return // The clicked cell is empty...
	}

	clicked := Point{row, col}

	if g.sel1 == nil { // First click
		g.sel1 = &clicked
		g.hint1 = nil
		g.hint2 = nil
		g.hintTTL = 0
	} else { // Second click
		if g.sel1.R == clicked.R && g.sel1.C == clicked.C {
			g.sel1 = nil // The same tile was clicked twice...
			return
		}

		path := FindPath(g.board, g.sel1.R, g.sel1.C, clicked.R, clicked.C)
		if path != nil { // Allowed move
			g.board.Set(g.sel1.R, g.sel1.C, TileEmpty)
			g.board.Set(clicked.R, clicked.C, TileEmpty)
			g.path = path
			g.pathTTL = 20
			g.sel1 = nil

			if g.board.RemainingTiles() == 0 {
				g.state = StateWin
				g.setMsg("You win! Press R to restart.", -1)
				return
			}

			if ok, _, _ := HasAnyMatch(g.board); !ok {
				if g.shuffles > 0 {
					g.setMsg("No moves available! Press S to shuffle.", 180)
				} else {
					g.state = StateStuck
					g.setMsg("No moves left! Press R to restart.", -1)
				}
			}
		} else { // Disallowed move
			g.sel1 = nil
			g.setMsg("No valid path or wrong kind!", 80)
		}
	}
}

func (g *Game) doHint() {
	ok, a, _ := HasAnyMatch(g.board)
	if ok {
		g.hint1 = &a
		g.hint2 = nil
		g.hintTTL = 90
	} else {
		g.setMsg("No matches available! Press S to shuffle.", 80)
	}
}

func (g *Game) doShuffle() {
	if g.shuffles <= 0 {
		g.setMsg("No shuffles remaining!", 80)
		return
	}

	g.shuffles--
	ShuffleRemaining(g.board, g.rng)
	g.sel1 = nil
	g.setMsg(fmt.Sprintf("Shuffled! (%d remaining)", g.shuffles), 80)
}

func (g *Game) setMsg(msg string, ttl int) {
	g.message = msg
	g.msgTTL = ttl
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(color.RGBA{0x2A, 0x2D, 0x35, 0xFF})

	// Draw board tiles
	r0, c0, r1, c1 := g.board.InnerBounds()
	for r := r0; r < r1; r++ {
		for c := c0; c < c1; c++ {
			kind := g.board.Get(r, c)
			if kind == TileEmpty {
				continue
			}

			x := g.offsetX + float64(c)*TileW
			y := g.offsetY + float64(r)*TileH
			selected := false

			// Highlight selected tiles
			if g.sel1 != nil && g.sel1.R == r && g.sel1.C == c {
				selected = true
			}

			// Highlight first hint tile
			if g.hintTTL > 0 {
				if (g.hint1 != nil && g.hint1.R == r && g.hint1.C == c) ||
					(g.hint2 != nil && g.hint2.R == r && g.hint2.C == c) {
					selected = true
				}
			}

			DrawTile(screen, kind, x, y, selected)
		}
	}

	// Draw path animation
	if g.pathTTL > 0 && g.path != nil {
		DrawPath(screen, g.path, g.offsetX, g.offsetY)
	}

	// Draw HUD
	g.drawHUD(screen)
}

func (g *Game) drawHUD(screen *ebiten.Image) {
	// Top bar message
	message := "Welcome to Rysz's Shisen-Go"
	msgColor := HudColorDefault

	if g.message != "" {
		message = g.message
		msgColor = HudColorHighlight
	}

	op := &text.DrawOptions{}
	w, h := text.Measure(message, g.uiFace, 0)
	op.GeoM.Translate(float64(ScreenW)/2-w/2, (float64(HudTopH-h)/2)+HudPadY)
	op.ColorScale.ScaleWithColor(msgColor)
	text.Draw(screen, message, g.uiFace, op)

	// Bottom bar message
	remaining := g.board.RemainingTiles()
	info := fmt.Sprintf("Tiles: %d  |  Shuffles: %d  |  [H] Hint  [S] Shuffle  [R] Restart  [A] Audio", remaining, g.shuffles)

	op = &text.DrawOptions{}
	w, h = text.Measure(info, g.uiFace, 0)
	op.GeoM.Translate(float64(ScreenW)/2-w/2, float64(ScreenH-HudBotH)+(float64(HudBotH)-h)/2-HudPadY)
	op.ColorScale.ScaleWithColor(HudColorDefault)
	text.Draw(screen, info, g.uiFace, op)
}

func main() {
	audioCtx := audio.NewContext(44100)

	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(musicOgg))
	if err != nil {
		log.Fatal(err)
	}
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	musicPlayer, err = audioCtx.NewPlayer(loop)
	if err != nil {
		log.Fatal(err)
	}
	musicPlayer.SetVolume(0.4)
	musicPlayer.Play()

	g := NewGame()
	g.initFonts()
	initTileCache(g.face)

	ebiten.SetWindowSize(ScreenW, ScreenH)
	ebiten.SetWindowTitle("Shisen-Sho")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
