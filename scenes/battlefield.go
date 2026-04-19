package scenes

import (
	"fmt"
	"image"
	"strconv"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/dqso/ludum-dare-59/player"
	"github.com/dqso/ludum-dare-59/token"
	"github.com/ebitengine/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

type battlefieldScene struct {
	game      entity.Game
	questions entity.QuestionsDatabase
	fontFace  font.Face

	player    entity.Player
	camera    Camera
	tokens    []entity.Token
	inventory [9]entity.CollectedToken

	tokenInFocus bool

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
			token.NewToken(fontFace, "не знаю", -100, 300, colornames.Peachpuff),
			token.NewToken(fontFace, "никогда", 100, 300, colornames.Tan),
			token.NewToken(fontFace, "две недели", 200, 250, colornames.Wheat),
		},
	}
}

func (s *battlefieldScene) Update() (entity.Scene, error) {
	if _, err := s.debug.Update(func(ctx *debugui.Context) error {
		const x, y = 10, 80
		ctx.Window("TODO", image.Rect(x, y, x+240, y+140), func(layout debugui.ContainerLayout) {
			ctx.Text(fmt.Sprintf("FPS: %0.2f; TPS: %0.2f", ebiten.ActualFPS(), ebiten.ActualTPS()))

			cx, cy := ebiten.CursorPosition()
			ctx.Text(fmt.Sprintf("Point Pos: (%d; %d)", cx, cy))
			ctx.Text(fmt.Sprintf("Player Pos: (%0.2f; %0.2f)", s.player.X(), s.player.Y()))

			ctx.Text("Press [Shift]+[F] for fullscreen.")
			if s.tokenInFocus {
				ctx.Text("You can pick up this phrase with [E].")
			}
		})
		return nil
	}); err != nil {
		return NewErrorScene(s.game, err), nil
	}

	if ebiten.IsKeyPressed(ebiten.KeyF) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
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

	s.tokenInFocus = false
	radius := s.player.PivotRadius()
	nearestIdx, nearestDistance := -1, 999.0
	for idx, t := range s.tokens {
		px, py := s.player.PivotX(), s.player.PivotY()
		cx := clamp(px, t.TopLeftX(), t.TopLeftX()+t.Width())
		cy := clamp(py, t.TopLeftY(), t.TopLeftY()+t.Height())
		dx := px - cx
		dy := py - cy
		distance2 := dx*dx + dy*dy
		if distance2 <= radius*radius {
			t.SetFocus(true)
			if distance2 < nearestDistance {
				nearestIdx, nearestDistance = idx, distance2
			}
			s.tokenInFocus = true
		} else {
			t.SetFocus(false)
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyE) && nearestIdx >= 0 {
		for idx := range s.inventory {
			if s.inventory[idx] != nil {
				continue
			}
			s.inventory[idx] = s.tokens[nearestIdx].Collect()
			s.tokens = append(s.tokens[:nearestIdx], s.tokens[nearestIdx+1:]...)
			break
		}
	}

	slotKeys := []ebiten.Key{
		ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5,
		ebiten.Key6, ebiten.Key7, ebiten.Key8, ebiten.Key9,
	}
	for _, slotKey := range slotKeys {
		if ebiten.IsKeyPressed(slotKey) {
			idx := slotKey - ebiten.Key1
			if s.inventory[idx] == nil {
				continue
			}
			t := s.inventory[idx].Drop(s.player.X(), s.player.Y())
			s.inventory[idx] = nil
			s.tokens = append(s.tokens, t)
			break
		}
	}

	return nil, nil
}

func (s *battlefieldScene) Draw(screen *ebiten.Image) {
	winW, _ := s.game.WindowSize()

	for _, t := range s.tokens {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(t.TopLeftX()-s.camera.X, t.TopLeftY()-s.camera.Y)
		t.Draw(screen, op)
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.player.TopLeftX()-s.camera.X, s.player.TopLeftY()-s.camera.Y)
	s.player.Draw(screen, op)

	slotWidth := float64(winW) / float64(len(s.inventory))
	for idx, ct := range s.inventory {
		if ct == nil {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		coeff := slotWidth / ct.Width()
		var dx float64
		if coeff < 1 {
			op.GeoM.Scale(coeff, coeff)
		} else {
			dx = slotWidth/2 - ct.Width()/2
		}
		op.GeoM.Translate(float64(idx)*slotWidth+dx, 0)
		ct.Draw(screen, op)

		label := strconv.Itoa(idx + 1)
		rect := text.BoundString(s.fontFace, label)
		sprite := ebiten.NewImage(rect.Dx(), rect.Dy())
		op.GeoM.Reset()
		op.ColorM.ScaleWithColor(colornames.Gray)
		op.GeoM.Translate(float64(-rect.Min.X), float64(-rect.Min.Y))
		text.DrawWithOptions(sprite, label, s.fontFace, op)

		op.GeoM.Reset()
		op.GeoM.Translate(
			float64(idx)*slotWidth+slotWidth/2-float64(rect.Dx())/2,
			ct.Height()+5,
		)
		screen.DrawImage(sprite, op)
	}

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
