package main

import (
	_ "embed"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"golang.org/x/image/font/gofont/goregular"
)

//go:embed music.ogg
var musicOgg []byte

// All the different states that a game can have.
const (
	statePlaying = iota
	stateWin
	stateStuck
)

const (
	screenW = 960 // Pixel width total of screen
	screenH = 650 // Pixel height total of screen

	tileW = 48 // Pixel width of a single tile
	tileH = 64 // Pixel height of a single tile

	hudTopH = 32 // Pixel height of the top HUD
	hudBotH = 32 // Pixel height of the bottom HUD
	hudPadY = 4  // Pixel height of HUD padding (Y)

	fontSizeHud  = 18
	fontSizeTile = 28
)

// Requirement: numTileKinds * tilesPerKind == innerRows * innerCols.
// Beware if this requirement is not met, the program will panic upon game creation.
const (
	numTileKinds = 36
	tilesPerKind = 4
	innerRows    = 8
	innerCols    = 18

	defaultShuffles = 3

	timeoutMessageNever   = -1 // frames
	timeoutMessageWarning = 80 // frames
	timeoutPathVisible    = 20 // frames
	timeoutHintVisible    = 90 // frames
)

var (
	GameName    = "Shisen-Go"
	GameVersion string

	_ ebiten.Game = (*Game)(nil)

	gameFont            = goregular.TTF
	gameColorBackground = color.RGBA{0x2A, 0x2D, 0x35, 0xFF}

	gameMusicPlayer     *audio.Player
	gameMusicSampleRate = 44100
	gameMusicVolume     = 0.4

	tileImageCache    map[TileKind]*ebiten.Image
	tileSelectedCache map[TileKind]*ebiten.Image

	tileColorBackground         = color.RGBA{0xF5, 0xF0, 0xE0, 0xFF} // cream
	tileColorBackgroundSelected = color.RGBA{0xFF, 0xFF, 0x80, 0xFF} // bright yellow
	tileColorShadow             = color.RGBA{0x80, 0x78, 0x68, 0xFF} // gray
	tileColorBorder             = color.RGBA{0x60, 0x58, 0x48, 0xFF} // brown
	tileColorBorderSelected     = color.RGBA{0xE0, 0xA0, 0x00, 0xFF} // bright yellow

	pathColor = color.RGBA{0xFF, 0xC0, 0x00, 0xFF} // gold

	hudIdleMessage         = "Welcome to Rysz's Shisen-Go"
	hudColorMessageDefault = color.RGBA{0x60, 0x60, 0x60, 0xFF}
	hudColorMessageWarning = color.RGBA{0xFF, 0xFF, 0x80, 0xFF}
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
