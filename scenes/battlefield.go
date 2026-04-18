package scenes

import (
	"fmt"
	"image"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/dqso/ludum-dare-59/player"
	"github.com/dqso/ludum-dare-59/token"
	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
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
	tokens []entity.Token

	debug debugui.DebugUI
}

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
		tokens: []entity.Token{
			token.NewToken(fontFace, "Any other offers?", -10, -65, colornames.Blue),
			token.NewToken(fontFace, "depends", -45, 115, colornames.Green),
			token.NewToken(fontFace, "no idea", 90, 70, colornames.Violet),
			token.NewToken(fontFace, "doesn't\nmatter", 150, -70, colornames.Navy),
			token.NewToken(fontFace, "When can you start?", 0, 200, colornames.Lightgray),
			token.NewToken(fontFace, "it's trendy", 0, -200, colornames.Darkgray),
			token.NewToken(fontFace, "Привет :)", 0, -300, colornames.Greenyellow),
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
			ctx.Text(fmt.Sprintf("Player Position: (%0.2f; %0.2f)", s.player.X(), s.player.Y()))
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

	radius := s.player.PivotRadius()
	for _, t := range s.tokens {
		px, py := s.player.PivotX(), s.player.PivotY()
		cx := clamp(px, t.TopLeftX(), t.TopLeftX()+t.Width())
		cy := clamp(py, t.TopLeftY(), t.TopLeftY()+t.Height())
		dx := px - cx
		dy := py - cy
		t.SetFocus(dx*dx+dy*dy <= radius*radius)
	}

	return nil, nil
}

func (s *battlefieldScene) Draw(screen *ebiten.Image) {
	//winW, winH := s.game.WindowSize()

	for _, token := range s.tokens {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(token.TopLeftX()-s.camera.X, token.TopLeftY()-s.camera.Y)
		token.Draw(screen, op)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.player.TopLeftX()-s.camera.X, s.player.TopLeftY()-s.camera.Y)
	s.player.Draw(screen, op)

	s.debug.Draw(screen)
}

func (s *battlefieldScene) Name() string { return "Battlefield" }

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
