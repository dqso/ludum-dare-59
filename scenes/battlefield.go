package scenes

import (
	"fmt"
	"image"
	"image/color"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/dqso/ludum-dare-59/player"
	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type battlefieldScene struct {
	game      entity.Game
	questions entity.QuestionsDatabase
	fontFace  font.Face

	player entity.Player
	camera Camera
	tokens []Token

	debug debugui.DebugUI
}

type Token struct {
	X, Y   float64
	W, H   float64
	Text   string
	Sprite *ebiten.Image
}

func NewToken(fontFace font.Face, str string, x, y float64, clr color.Color) Token {
	rect := text.BoundString(fontFace, str)
	const padding = 5
	sprite := ebiten.NewImage(rect.Dx()+padding*2, rect.Dy()+padding*2)
	r32, g32, b32, a32 := clr.RGBA()
	r, g, b, a := uint8(r32>>8), uint8(g32>>8), uint8(b32>>8), uint8(a32>>8)
	//sprite.Fill(color.RGBA{
	//	R: r + (255-r)/2,
	//	G: g + (255-g)/2,
	//	B: b + (255-b)/2,
	//	A: a,
	//})
	_ = a
	luma := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b) // яркость по формуле luma
	var shift float64
	if luma > 128 {
		shift = -150 // светлый текст → тёмный
	} else {
		shift = +150 // тёмный текст → светлый фон
	}
	clamp := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	sprite.Fill(color.RGBA{
		R: clamp(float64(r) + shift),
		G: clamp(float64(g) + shift),
		B: clamp(float64(b) + shift),
		A: 200,
	})

	vector.StrokeRect(sprite, 0, 0, float32(rect.Dx()-1)+padding*2, float32(rect.Dy()-1)+padding*2, 1, clr, false)
	var opts ebiten.DrawImageOptions
	opts.ColorM.ScaleWithColor(clr)
	opts.GeoM.Translate(float64(-rect.Min.X)+padding, float64(-rect.Min.Y)+padding)
	text.DrawWithOptions(sprite, str, fontFace, &opts)

	return Token{
		X:      x,
		Y:      y,
		W:      float64(rect.Dx()) + padding*2,
		H:      float64(rect.Dy()) + padding*2,
		Text:   str,
		Sprite: sprite,
	}
}

func (t Token) DrawX() float64 { return t.X - t.W/2 }
func (t Token) DrawY() float64 { return t.Y - t.H/2 }

type Camera struct {
	X, Y float64
}

func (c *Camera) Follow(target entity.Positionable, winW, winH float64) {
	if x := target.X() - c.X; x < winW/4 {
		c.X = target.X() - winW/4
	} else if x > winW*3/4 {
		c.X = target.X() - winW*3/4
	}
	if y := target.Y() - c.Y; y < winH/4 {
		c.Y = target.Y() - winH/4
	} else if y > winH*3/4 {
		c.Y = target.Y() - winH*3/4
	}
}

func NewBattlefieldScene(game entity.Game, questions entity.QuestionsDatabase) entity.Scene {
	winW, winH := game.WindowSize()

	player, err := player.NewPlayer(0, 0)
	if err != nil {
		return NewErrorScene(game, err)
	}
	camera := Camera{
		X: player.X() - float64(winW)/2,
		Y: player.Y() - float64(winH)/2,
	}

	tt, err := opentype.Parse(assets.StampatelloFacetoKernTTF)
	if err != nil {
		return NewErrorScene(game, err)
	}
	fontFace, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    float64(14),
		DPI:     96,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return NewErrorScene(game, err)
	}

	return &battlefieldScene{
		game:      game,
		questions: questions,
		fontFace:  fontFace,

		player: player,
		camera: camera,
		tokens: []Token{
			NewToken(fontFace, "Any other offers?", -10, -65, colornames.Blue),
			NewToken(fontFace, "depends", -45, 115, colornames.Green),
			NewToken(fontFace, "no idea", 90, 70, colornames.Violet),
			NewToken(fontFace, "doesn't\nmatter", 150, -70, colornames.Navy),
			NewToken(fontFace, "When can you start?", 0, 200, colornames.Lightgray),
			NewToken(fontFace, "it's trendy", 0, -200, colornames.Darkgray),
			NewToken(fontFace, "Привет :)", 0, -300, colornames.Greenyellow),
		},
	}
}

func (s *battlefieldScene) Update() (entity.Scene, error) {
	if _, err := s.debug.Update(func(ctx *debugui.Context) error {
		ctx.Window("TODO", image.Rect(10, 10, 260, 160), func(layout debugui.ContainerLayout) {
			ctx.Text(fmt.Sprintf("FPS: %0.2f", ebiten.ActualFPS()))
			ctx.Text(fmt.Sprintf("TPS: %0.2f", ebiten.ActualTPS()))

			cx, cy := ebiten.CursorPosition()
			ctx.Text(fmt.Sprintf("Point: (%d; %d)", cx, cy))

			ctx.Text(fmt.Sprintf("Camera Position: (%0.2f; %0.2f)", s.camera.X, s.camera.Y))
			ctx.Text(fmt.Sprintf("World Position: (%0.2f; %0.2f)", s.player.X(), s.player.Y()))
		})
		return nil
	}); err != nil {
		return NewErrorScene(s.game, err), nil
	}

	step := s.player.Speed()
	dx, dy := 0.0, 0.0
	if ebiten.IsKeyPressed(ebiten.KeyLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx -= step
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx += step
	}
	if ebiten.IsKeyPressed(ebiten.KeyDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy += step
	}
	if ebiten.IsKeyPressed(ebiten.KeyUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy -= step
	}
	s.player.Move(dx, dy)

	winW, winH := s.game.WindowSize()
	_ = winH

	s.camera.Follow(s.player, float64(winW), float64(winH))

	return nil, nil
}

func (s *battlefieldScene) Draw(screen *ebiten.Image) {
	//winW, winH := s.game.WindowSize()

	for _, token := range s.tokens {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(token.DrawX()-s.camera.X, token.DrawY()-s.camera.Y)
		screen.DrawImage(token.Sprite, op)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.player.TopLeftX()-s.camera.X, s.player.TopLeftY()-s.camera.Y)
	s.player.Draw(screen, op)

	s.debug.Draw(screen)
}

func (s *battlefieldScene) Name() string { return "Battlefield" }
