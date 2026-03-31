package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type hudButton struct {
	x0, x1 int // pixels
	action string
}

// Game is the primary structure for our game.
type Game struct {
	board *Board
	rng   *rand.Rand

	state    int // statePlaying, stateWin, stateStuck
	shuffles int // Remaining shuffles for the game

	audioMuted bool // If audio was muted by user

	hudFace  text.Face // Font face for UI text
	tileFace text.Face // Font face for tile symbols

	hudButtons []hudButton // HUD button bounds

	offsetX float64 // Pixel offset X to center board
	offsetY float64 // Pixel offset Y to center board

	sel1    *Point  // First selected tile (nil if none)
	path    []Point // Last matched path (for animation)
	pathTTL int     // Ticks remaining to show the path

	hint1   *Point // Hint tile 1
	hint2   *Point // Hint tile 2
	hintTTL int    // Ticks remaining to show the hints

	message string // Message (shuffled, stuck, etc...)
	msgTTL  int    // Ticks remaining to show the message
}

// NewGame returns a new Game instance, can be called multiple times.
func NewGame() *Game {
	rng := rand.New(rand.NewSource(time.Now().UnixNano())) //nolint:gosec
	board := generateSolvableBoard(gameInnerRows, gameInnerCols, gameNumSymbols, gameTilesPerSymbol, rng)

	g := &Game{
		board:    board,
		rng:      rng,
		state:    statePlaying,
		shuffles: gameShuffles,
	}

	// Font
	src, err := text.NewGoTextFaceSource(bytes.NewReader(gameFont))
	if err != nil {
		log.Fatal(err)
	}

	// HUD
	g.hudFace = &text.GoTextFace{
		Source: src,
		Size:   fontSizeHud,
	}

	// Tiles
	g.tileFace = &text.GoTextFace{
		Source: src,
		Size:   fontSizeTile,
	}

	// Calculate offsets to center the board
	boardPixelW := float64(board.Cols) * tileW
	boardPixelH := float64(board.Rows) * tileH

	g.offsetX = (screenW - boardPixelW) / 2
	g.offsetY = hudTopH + (screenH-hudTopH-hudBotH-boardPixelH)/2

	return g
}

// Update handles the updating of all game elements and also keystrokes.
//
//nolint:gocognit,cyclop,funlen
func (g *Game) Update() error {
	// Update HUD
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

	// Register click events (mouse or touches)
	clicks := []struct{ x, y int }{}
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		tx, ty := ebiten.TouchPosition(id)
		clicks = append(clicks, struct{ x, y int }{tx, ty})
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		mx, my := ebiten.CursorPosition()
		clicks = append(clicks, struct{ x, y int }{mx, my})
	}

	if g.state == statePlaying {
		// Hint a possible pairing
		if inpututil.IsKeyJustPressed(ebiten.KeyH) {
			g.doHint()
		}
		// Shuffle remaining tiles
		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.doShuffle()
		}
	}
	// Toggle the audio
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		g.audioMuted = !g.audioMuted
		if g.audioMuted {
			gameMusicPlayer.Pause()
		} else {
			gameMusicPlayer.Play()
		}
	}
	// Restart the game
	if inpututil.IsKeyJustPressed(ebiten.KeyR) {
		*g = *NewGame()

		return nil
	}

	// Process click events - HUD first, then tile selection:
	g.computeHudButtons()
	for _, click := range clicks {
		if click.y >= screenH-hudBotH {
			for _, btn := range g.hudButtons {
				if click.x >= btn.x0 && click.x < btn.x1 {
					switch btn.action {
					case "hint":
						if g.state == statePlaying {
							g.doHint()
						}
					case "shuffle":
						if g.state == statePlaying {
							g.doShuffle()
						}
					case "restart":
						*g = *NewGame()

						return nil
					case "audio":
						g.audioMuted = !g.audioMuted
						if g.audioMuted {
							gameMusicPlayer.Pause()
						} else {
							gameMusicPlayer.Play()
						}
					}

					break
				}
			}
		} else if g.state == statePlaying {
			g.handleClick(click.x, click.y)
		}
	}

	return nil
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenW, screenH
}

// Draw handles the drawing of all game graphics.
func (g *Game) Draw(screen *ebiten.Image) {
	// Background
	screen.Fill(gameColorBackground)

	// Draw board tiles
	r0, c0, r1, c1 := g.board.InnerBounds()
	for r := r0; r < r1; r++ {
		for c := c0; c < c1; c++ {
			sym := g.board.Get(r, c)
			if sym == TileEmpty {
				continue
			}

			x := g.offsetX + float64(c)*tileW
			y := g.offsetY + float64(r)*tileH
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

			drawTile(screen, sym, x, y, selected)
		}
	}

	// Draw path animation
	if g.pathTTL > 0 && g.path != nil {
		drawPath(screen, g.path, g.offsetX, g.offsetY)
	}

	// Draw HUD
	g.drawHUD(screen)
}

// drawHUD is a helper function that draws the HUD onto the screen.
func (g *Game) drawHUD(screen *ebiten.Image) {
	// Draw the HUD at the top of the screen:
	msg := hudIdleMessage
	msgColor := hudColorMessageDefault

	if g.message != "" {
		msg = g.message
		msgColor = hudColorMessageWarning
	}

	op := &text.DrawOptions{}
	w, h := text.Measure(msg, g.hudFace, 0)
	op.GeoM.Translate(float64(screenW)/2-w/2, (float64(hudTopH-h)/2)+hudPadY)
	op.ColorScale.ScaleWithColor(msgColor)
	text.Draw(screen, msg, g.hudFace, op)

	// Draw the HUD at the bottom of the screen:
	remaining := g.board.RemainingTiles()
	info := fmt.Sprintf("Tiles: %d  |  Shuffles: %d  |  [H] Hint  [S] Shuffle  [R] Restart  [A] Audio", remaining, g.shuffles)

	op = &text.DrawOptions{}
	w, h = text.Measure(info, g.hudFace, 0)
	op.GeoM.Translate(float64(screenW)/2-w/2, float64(screenH-hudBotH)+(float64(hudBotH)-h)/2-hudPadY)
	op.ColorScale.ScaleWithColor(hudColorMessageDefault)
	text.Draw(screen, info, g.hudFace, op)
}

// Compute HUD button zones from actual text measurements.
func (g *Game) computeHudButtons() {
	prefix := fmt.Sprintf("Tiles: %d  |  Shuffles: %d  |  ", g.board.RemainingTiles(), g.shuffles)

	actions := []struct {
		label  string
		action string
	}{
		{"[H] Hint  ", "hint"},
		{"[S] Shuffle  ", "shuffle"},
		{"[R] Restart  ", "restart"},
		{"[A] Audio", "audio"},
	}

	fullStr := prefix
	pw, _ := text.Measure(prefix, g.hudFace, 0)
	hudW, _ := text.Measure(fullStr+actions[0].label+actions[1].label+actions[2].label+actions[3].label, g.hudFace, 0)
	baseX := (float64(screenW) - hudW) / 2

	g.hudButtons = nil
	offsetX := baseX + pw

	for _, a := range actions {
		w, _ := text.Measure(a.label, g.hudFace, 0)
		g.hudButtons = append(g.hudButtons, hudButton{
			x0:     int(offsetX),
			x1:     int(offsetX + w),
			action: a.action,
		})
		offsetX += w
	}
}

// handeClick implements the clicking logic for selecting tiles.
func (g *Game) handleClick(mx, my int) {
	// Convert pixel to board coordinates
	col := int(float64(mx)-g.offsetX) / tileW
	row := int(float64(my)-g.offsetY) / tileH
	clicked := Point{row, col}

	// Check if click is within playable area
	r0, c0, r1, c1 := g.board.InnerBounds()
	if col < 0 || row < 0 || row < r0 || row >= r1 || col < c0 || col >= c1 {
		return
	}
	if g.board.IsEmpty(row, col) {
		return
	}

	// If it's the first click, register and bail out.
	if g.sel1 == nil {
		g.sel1 = &clicked
		g.hint1 = nil
		g.hint2 = nil
		g.hintTTL = 0

		return
	}

	// The second click was the same as the first click (deselect).
	if g.sel1.R == clicked.R && g.sel1.C == clicked.C {
		g.sel1 = nil

		return
	}

	// Check if it's a valid connection between the selected tiles.
	path := findPath(g.board, g.sel1.R, g.sel1.C, clicked.R, clicked.C)
	if path != nil {
		g.board.Set(g.sel1.R, g.sel1.C, TileEmpty)
		g.board.Set(clicked.R, clicked.C, TileEmpty)

		g.path = path
		g.pathTTL = timeoutPathVisible
		g.sel1 = nil

		if g.board.RemainingTiles() == 0 {
			g.state = stateWin
			g.setMessage("You win! Press R to restart.", timeoutMessageNever)

			return
		}

		if ok, _, _ := hasAnyMatch(g.board); !ok {
			if g.shuffles > 0 {
				g.setMessage("No moves available! Press S to shuffle.", timeoutMessageNever)
			} else {
				g.state = stateStuck
				g.setMessage("No moves left! Press R to restart.", timeoutMessageNever)
			}
		}

		return
	}

	// No connection is possible.
	g.sel1 = nil
	g.setMessage("No valid path or symbol not matching!", timeoutMessageWarning)
}

// setMessage sets a message and TTL to be shown in the HUD.
func (g *Game) setMessage(msg string, ttl int) {
	g.message = msg
	g.msgTTL = ttl
}

// doHint handles when the users requests a hint to be shown.
func (g *Game) doHint() {
	// Check if we have any possible connections left.
	ok, a, _ := hasAnyMatch(g.board)
	if ok {
		g.hint1 = &a
		g.hint2 = nil
		g.hintTTL = timeoutHintVisible

		return
	}

	// None are left, offer to shuffle or restart the game (if no shuffles left).
	if g.shuffles > 0 {
		g.setMessage("No moves available! Press S to shuffle.", timeoutMessageNever)
	} else {
		g.state = stateStuck
		g.setMessage("No moves left! Press R to restart.", timeoutMessageNever)
	}
}

func (g *Game) doShuffle() {
	// Check if we have shuffles left, otherwise bail out.
	if g.shuffles <= 0 {
		g.setMessage("No shuffles remaining!", timeoutMessageWarning)

		return
	}

	// Reduce the shuffles and shuffle the remaining tiles.
	g.shuffles--
	shuffleRemaining(g.board, g.rng)

	// Deselect the selected tile as the position has changed.
	g.sel1 = nil
	g.setMessage(fmt.Sprintf("Shuffled! (%d remaining)", g.shuffles), timeoutMessageWarning)
}

func main() {
	audioCtx := audio.NewContext(gameMusicSampleRate)
	stream, err := vorbis.DecodeWithoutResampling(bytes.NewReader(musicOgg))
	if err != nil {
		log.Fatal(err)
	}
	loop := audio.NewInfiniteLoop(stream, stream.Length())
	gameMusicPlayer, err = audioCtx.NewPlayer(loop)
	if err != nil {
		log.Fatal(err)
	}
	gameMusicPlayer.SetVolume(gameMusicVolume)
	gameMusicPlayer.Play()

	g := NewGame()
	initTileCache(g.tileFace)

	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle(GameName + " " + GameVersion)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(g); err != nil {
		log.Fatal(err)
	}
}
