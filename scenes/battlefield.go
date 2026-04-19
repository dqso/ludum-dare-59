package scenes

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"
	"strconv"
	"time"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/character"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/dqso/ludum-dare-59/player"
	"github.com/dqso/ludum-dare-59/token"
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

	gameStart            time.Time
	nextTokenSpawn       time.Time
	nextInterviewerSpawn time.Time

	player     entity.Player
	camera     Camera
	characters []entity.Character
	tokens     []entity.Token
	inventory  [9]entity.CollectedToken

	nearestCharacterIdx int
	tokensInFocus       int

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

const (
	maxSpawnDistance = 500.0

	tokenPoolSize       = 10
	tokenSpawnBatchSize = 3
	tokenSpawnDelayFrom = time.Second * 20
	tokenSpawnDelayTo   = time.Second * 30
	tokenLifetimeFrom   = time.Second * 50
	tokenLifetimeTo     = time.Second * 70

	interviewerRecruiterPoolSize = 4
	interviewerSpawnBatchSize    = 1
	interviewerSpawnDelayFrom    = time.Second * 60
	interviewerSpawnDelayTo      = time.Second * 90
	interviewerLifetimeFrom      = time.Minute * 3
	interviewerLifetimeTo        = time.Minute * 4
)

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

	now := time.Now()
	return &battlefieldScene{
		game:      game,
		questions: questions,
		fontFace:  fontFace,

		gameStart:            now,
		nextTokenSpawn:       now,
		nextInterviewerSpawn: now,

		player:     player,
		camera:     camera,
		characters: make([]entity.Character, 0),
		tokens:     make([]entity.Token, 0),
	}
}

func (s *battlefieldScene) Update() (entity.Scene, error) {
	for i := len(s.characters) - 1; i >= 0; i-- {
		if time.Since(s.characters[i].Deadline()) > 0 {
			s.characters = append(s.characters[:i], s.characters[i+1:]...)
		}
	}
	for i := len(s.tokens) - 1; i >= 0; i-- {
		if time.Since(s.tokens[i].Deadline()) > 0 {
			s.tokens = append(s.tokens[:i], s.tokens[i+1:]...)
		}
	}
	if time.Since(s.nextTokenSpawn) > 0 {
		s.nextTokenSpawn = time.Now().Add(randDuration(tokenSpawnDelayFrom, tokenSpawnDelayTo))
		tokensToSpawn := min(
			max(tokenPoolSize-len(s.tokens), 0),
			tokenSpawnBatchSize,
		)
		for _, a := range s.questions.GetRandomAnswers(tokensToSpawn) {
			t := token.NewToken(s.fontFace, a,
				float64(rand.IntN(maxSpawnDistance*2)-maxSpawnDistance),
				float64(rand.IntN(maxSpawnDistance*2)-maxSpawnDistance),
				randDuration(tokenLifetimeFrom, tokenLifetimeTo),
			)
			s.tokens = append(s.tokens, t)
		}
	}
	if time.Since(s.nextInterviewerSpawn) > 0 {
		s.nextInterviewerSpawn = time.Now().Add(randDuration(interviewerSpawnDelayFrom, interviewerSpawnDelayTo))
		var numRecruiters int
		for _, c := range s.characters {
			if c.Role() == entity.CharacterRoleRecruiter {
				numRecruiters++
			}
		}
		interviewersToSpawn := min(
			max(interviewerRecruiterPoolSize-numRecruiters, 0),
			interviewerSpawnBatchSize,
		)
		for range interviewersToSpawn {
			c, err := character.NewCharacter(entity.CharacterRoleRecruiter,
				float64(rand.IntN(maxSpawnDistance*2)-maxSpawnDistance),
				float64(rand.IntN(maxSpawnDistance*2)-maxSpawnDistance),
				s.questions.GetRandomQuestions(3),
				randDuration(interviewerLifetimeFrom, interviewerLifetimeTo),
			)
			if err != nil {
				return NewErrorScene(s.game, err), nil
			}
			s.characters = append(s.characters, c)
		}
	}
	if _, err := s.debug.Update(func(ctx *debugui.Context) error {
		const x, y = 10, 80
		ctx.Window("TODO", image.Rect(x, y, x+250, y+140), func(layout debugui.ContainerLayout) {
			ctx.Text(fmt.Sprintf("FPS: %0.2f; TPS: %0.2f", ebiten.ActualFPS(), ebiten.ActualTPS()))
			ctx.Text(fmt.Sprintf("%s", time.Now().Format(time.Kitchen)))

			//cx, cy := ebiten.CursorPosition()
			//ctx.Text(fmt.Sprintf("Point Pos: (%d; %d)", cx, cy))
			//ctx.Text(fmt.Sprintf("Player Pos: (%0.2f; %0.2f)", s.player.X(), s.player.Y()))

			// TODO убирать про фулскрин, после полминуты игры
			ctx.Text("Use [Shift]+[F] for fullscreen.")
			if s.tokensInFocus > 0 {
				ctx.Text("Use [E] to pick up phrases.")
			}
			if s.nearestCharacterIdx >= 0 {
				ctx.Text("Use [1]-[9] to answer.")
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

	s.tokensInFocus = 0
	playerRadius := s.player.PivotRadius()
	nearestTokenIdx, nearestTokenDistance := -1, 99999.0
	for idx, t := range s.tokens {
		px, py := s.player.PivotX(), s.player.PivotY()
		cx := clamp(px, t.TopLeftX(), t.TopLeftX()+t.Width())
		cy := clamp(py, t.TopLeftY(), t.TopLeftY()+t.Height())
		dx := px - cx
		dy := py - cy
		distance2 := dx*dx + dy*dy
		if distance2 <= playerRadius*playerRadius {
			t.SetFocus(true)
			if distance2 < nearestTokenDistance {
				nearestTokenIdx, nearestTokenDistance = idx, distance2
			}
			s.tokensInFocus++
		} else {
			t.SetFocus(false)
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyE) && nearestTokenIdx >= 0 {
		for idx := range s.inventory {
			if s.inventory[idx] != nil {
				continue
			}
			s.inventory[idx] = s.tokens[nearestTokenIdx].Collect()
			s.tokens = append(s.tokens[:nearestTokenIdx], s.tokens[nearestTokenIdx+1:]...)
			break
		}
	}

	s.nearestCharacterIdx = -1
	nearestCharacterDistance := 99999.0
	const contactDeltaRadius = 15.0
	for idx, c := range s.characters {
		px, py := s.player.PivotX(), s.player.PivotY()
		dx := px - c.PivotX()
		dy := py - c.PivotY()
		distance2 := dx*dx + dy*dy
		touchRadius := playerRadius + c.PivotRadius()
		if distance2 < (touchRadius+contactDeltaRadius)*(touchRadius+contactDeltaRadius) {
			c.SetFocus(true)
			if distance2 < nearestCharacterDistance {
				s.nearestCharacterIdx = idx
				nearestCharacterDistance = distance2
			}
			if distance2 < touchRadius*touchRadius && distance2 > 0 {
				dist := math.Sqrt(distance2)
				push := touchRadius - dist
				s.player.Move(dx/dist*push, dy/dist*push)
			}
		} else {
			c.SetFocus(false)
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

			if s.nearestCharacterIdx >= 0 {
				answer := t.Answer()
				s.characters[s.nearestCharacterIdx].AnswerTheQuestion(s.questions, answer)
			} else {
				s.tokens = append(s.tokens, t)
			}
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

	for _, c := range s.characters {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(c.TopLeftX()-s.camera.X, c.TopLeftY()-s.camera.Y)
		c.Draw(screen, op)
		if c.IsFocused() {
			if q := c.GetQuestion(); q != nil {
				tipX := c.TopLeftX() + c.Width()/2 - s.camera.X
				tipY := c.TopLeftY() - s.camera.Y
				drawSpeechBubble(screen, s.fontFace, q.Question(), tipX, tipY)
			}
		}
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

func drawSpeechBubble(screen *ebiten.Image, face font.Face, txt string, tipX, tipY float64) {
	const pad = 10.0
	const r = 8.0
	const arrowH = 14.0
	const arrowHalfW = 8.0

	bounds := text.BoundString(face, txt)
	w := float64(bounds.Dx()) + pad*2
	h := float64(bounds.Dy()) + pad*2

	x := tipX - w/2
	y := tipY - h - arrowH

	fillColor := func(vs []ebiten.Vertex) {
		for i := range vs {
			vs[i].SrcX = 1
			vs[i].SrcY = 1
			vs[i].ColorR = 0.98
			vs[i].ColorG = 0.98
			vs[i].ColorB = 0.9
			vs[i].ColorA = 0.95
		}
	}
	opts := &ebiten.DrawTrianglesOptions{AntiAlias: true}

	var body vector.Path
	body.MoveTo(float32(x+r), float32(y))
	body.LineTo(float32(x+w-r), float32(y))
	body.ArcTo(float32(x+w), float32(y), float32(x+w), float32(y+r), float32(r))
	body.LineTo(float32(x+w), float32(y+h-r))
	body.ArcTo(float32(x+w), float32(y+h), float32(x+w-r), float32(y+h), float32(r))
	body.LineTo(float32(x+r), float32(y+h))
	body.ArcTo(float32(x), float32(y+h), float32(x), float32(y+h-r), float32(r))
	body.LineTo(float32(x), float32(y+r))
	body.ArcTo(float32(x), float32(y), float32(x+r), float32(y), float32(r))
	body.Close()
	vs, is := body.AppendVerticesAndIndicesForFilling(nil, nil)
	fillColor(vs)
	screen.DrawTriangles(vs, is, whiteImage, opts)

	var arrow vector.Path
	arrow.MoveTo(float32(tipX-arrowHalfW), float32(y+h))
	arrow.LineTo(float32(tipX), float32(tipY))
	arrow.LineTo(float32(tipX+arrowHalfW), float32(y+h))
	arrow.Close()
	vs, is = arrow.AppendVerticesAndIndicesForFilling(nil, nil)
	fillColor(vs)
	screen.DrawTriangles(vs, is, whiteImage, opts)

	op := &ebiten.DrawImageOptions{}
	op.ColorM.ScaleWithColor(colornames.Black)
	op.GeoM.Translate(x+pad-float64(bounds.Min.X), y+pad-float64(bounds.Min.Y))
	text.DrawWithOptions(screen, txt, face, op)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

var whiteImage *ebiten.Image

func init() {
	whiteImage = ebiten.NewImage(3, 3)
	whiteImage.Fill(color.White)
}

func randDuration(from, to time.Duration) time.Duration {
	return from + time.Duration(rand.Int64N(int64(to-from)))
}
