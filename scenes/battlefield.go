package scenes

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"log"
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
	characters entity.CharacterList
	tokens     []entity.Token
	inventory  [9]entity.CollectedToken

	tokensInFocus    int
	offscreenSignals []offscreenSignal

	debug debugui.DebugUI
}

type offscreenSignal struct {
	color      color.Color
	signalType SignalType
	x, y       float64
	angle      float64
}

type SignalType uint8

const (
	SignalTypeToken SignalType = iota
	SignalTypeInterviewer
)

type Camera struct {
	X, Y float64
}

const (
	coeffScreenXToMove = 0.4
	coeffScreenYToMove = 0.4
)

func (c *Camera) Follow(target entity.Positionable, winW, winH float64) {
	if x := target.X() - c.X; x < winW*coeffScreenXToMove {
		c.X = target.X() - winW*coeffScreenXToMove
	} else if x > winW*(1-coeffScreenXToMove) {
		c.X = target.X() - winW*(1-coeffScreenXToMove)
	}
	if y := target.Y() - c.Y; y < winH*coeffScreenYToMove {
		c.Y = target.Y() - winH*coeffScreenYToMove
	} else if y > winH*(1-coeffScreenYToMove) {
		c.Y = target.Y() - winH*(1-coeffScreenYToMove)
	}
}

const (
	maxSpawnDistance = 500.0

	tokenPoolSize           = 20
	tokenSpawnBatchSizeFrom = 12
	tokenSpawnBatchSizeTo   = 15
	tokenSpawnDelayFrom     = time.Second * 20
	tokenSpawnDelayTo       = time.Second * 30
	tokenLifetimeFrom       = time.Second * 50
	tokenLifetimeTo         = time.Second * 70

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
		characters: character.NewList(),
		tokens:     make([]entity.Token, 0),
	}
}

func (s *battlefieldScene) Update() (entity.Scene, error) {
	for c, deleted := range s.characters.DeleteFunc(func(c entity.Character) bool {
		return time.Since(c.Deadline()) > 0
	}) {
		if deleted {
			log.Printf("Interviewer %s from company %q has been removed.", c.Role(), c.Company())
		}
	}
idleFor:
	for c := range s.characters.FilterFunc(character.FilterIdle) {
		var recruiter, engineer, founder entity.Character
		for e := range s.characters.FilterByCompany(c.Company()) {
			switch e.Role() {
			case entity.CharacterRoleRecruiter:
				recruiter = e
			case entity.CharacterRoleEngineer:
				engineer = e
			case entity.CharacterRoleFounder:
				founder = e
			default:
				continue idleFor
			}
		}
		spawnAngle := rand.Float64() * 2 * math.Pi
		spawnDist := recruiter.Width() * (2 + rand.Float64())
		spawnX, spawnY := recruiter.X()+math.Cos(spawnAngle)*spawnDist, recruiter.Y()+math.Sin(spawnAngle)*spawnDist
		if founder != nil && character.FilterIdle(founder) {
			if founder.GetQuestion() == nil {
				if founder.PlayerPoints() > 0 {
					founder.SetInterviewResult(character.NewRoundPassed())
				} else {
					founder.SetInterviewResult(character.NewRejection())
					break idleFor
				}
			}
			recruiter.SetInterviewResult(character.NewOffer(123))
		} else if engineer != nil && character.FilterIdle(engineer) {
			if engineer.GetQuestion() == nil {
				if engineer.PlayerPoints() > 0 {
					engineer.SetInterviewResult(character.NewRoundPassed())
				} else {
					engineer.SetInterviewResult(character.NewRejection())
					continue idleFor
				}
			}
			newChar, err := character.NewCharacter(
				entity.CharacterRoleFounder,
				spawnX, spawnY,
				s.questions.GetRandomQuestions(3),
				randDuration(interviewerLifetimeFrom, interviewerLifetimeTo),
				c.Company(),
			)
			if err != nil {
				return NewErrorScene(s.game, err), nil
			}
			s.characters.Add(newChar)
		} else if recruiter != nil && character.FilterIdle(recruiter) {
			if recruiter.GetQuestion() == nil {
				if recruiter.PlayerPoints() > 0 {
					recruiter.SetInterviewResult(character.NewRoundPassed())
				} else {
					recruiter.SetInterviewResult(character.NewRejection())
					continue idleFor
				}
			}
			newChar, err := character.NewCharacter(
				entity.CharacterRoleEngineer,
				spawnX, spawnY,
				s.questions.GetRandomQuestions(3),
				randDuration(interviewerLifetimeFrom, interviewerLifetimeTo),
				c.Company(),
			)
			if err != nil {
				return NewErrorScene(s.game, err), nil
			}
			s.characters.Add(newChar)
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
			tokenSpawnBatchSizeFrom+rand.IntN(tokenSpawnBatchSizeTo-tokenSpawnBatchSizeFrom),
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
		for range s.characters.FilterByRoles(entity.CharacterRoleRecruiter) {
			numRecruiters++
		}
		interviewersToSpawn := min(
			max(interviewerRecruiterPoolSize-numRecruiters, 0),
			interviewerSpawnBatchSize,
		)
		for range interviewersToSpawn {
			company := hex.EncodeToString(binary.LittleEndian.AppendUint64(nil, uint64(time.Now().UnixNano())))
			c, err := character.NewCharacter(entity.CharacterRoleRecruiter,
				float64(rand.IntN(maxSpawnDistance*2)-maxSpawnDistance),
				float64(rand.IntN(maxSpawnDistance*2)-maxSpawnDistance),
				s.questions.GetRandomQuestions(3),
				randDuration(interviewerLifetimeFrom, interviewerLifetimeTo),
				company,
			)
			if err != nil {
				return NewErrorScene(s.game, err), nil
			}
			s.characters.Add(c)
		}
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

	var nearestCharacter entity.Character
	nearestCharacterDistance := 99999.0
	const contactDeltaRadius = 15.0
	for c := range s.characters.All() {
		px, py := s.player.PivotX(), s.player.PivotY()
		dx := px - c.PivotX()
		dy := py - c.PivotY()
		distance2 := dx*dx + dy*dy
		touchRadius := playerRadius + c.PivotRadius()
		if distance2 < (touchRadius+contactDeltaRadius)*(touchRadius+contactDeltaRadius) {
			c.SetFocus(true)
			if distance2 < nearestCharacterDistance {
				nearestCharacter = c
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

			isGiveToCharacter := nearestCharacter != nil &&
				(nearestCharacter.GetQuestion() != nil ||
					(nearestCharacter.InterviewResult() != nil && nearestCharacter.InterviewResult().Outcome() == entity.OutcomeOffer))
			if isGiveToCharacter {
				answer := t.Answer()
				nearestCharacter.AnswerTheQuestion(s.questions, answer)
			} else {
				s.tokens = append(s.tokens, t)
			}
			break
		}
	}

	winW64, winH64 := float64(winW), float64(winH)
	s.offscreenSignals = s.offscreenSignals[:0]
	newOffscreenSignal := func(obj interface {
		entity.Positionable
		entity.TopLeftGetter
		entity.SizeGetter
	}, signalType SignalType, clr color.Color) (offscreenSignal, bool) {
		if obj.TopLeftX() < s.camera.X+winW64 && obj.TopLeftX()+obj.Width() > s.camera.X &&
			obj.TopLeftY() < s.camera.Y+winH64 && obj.TopLeftY()+obj.Height() > s.camera.Y {
			return offscreenSignal{}, false
		}
		tx := obj.X() - s.camera.X
		ty := obj.Y() - s.camera.Y
		cx, cy := winW64/2, winH64/2
		dx, dy := tx-cx, ty-cy
		const margin = 10.0
		scaleX := (winW64/2 - margin) / math.Abs(dx)
		scaleY := (winH64/2 - margin) / math.Abs(dy)
		sc := math.Min(scaleX, scaleY)
		return offscreenSignal{
			color:      clr,
			signalType: signalType,
			x:          cx + dx*sc,
			y:          cy + dy*sc,
			angle:      math.Atan2(dy, dx),
		}, true
	}
	for _, t := range s.tokens {
		sig, ok := newOffscreenSignal(t, SignalTypeToken, entity.AnswerCategoryToColor(t.Answer().Category()))
		if ok {
			s.offscreenSignals = append(s.offscreenSignals, sig)
		}
	}
	for c := range s.characters.All() {
		sig, ok := newOffscreenSignal(c, SignalTypeInterviewer, colornames.Darkred)
		if ok {
			s.offscreenSignals = append(s.offscreenSignals, sig)
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
			if nearestCharacter != nil {
				ctx.Text("Use [1]-[9] to answer.")
			}
		})
		return nil
	}); err != nil {
		return NewErrorScene(s.game, err), nil
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

	for c := range s.characters.All() {
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(c.TopLeftX()-s.camera.X, c.TopLeftY()-s.camera.Y)
		c.Draw(screen, op)
	}
	for c := range s.characters.All() {
		if c.IsFocused() {
			if q := c.GetQuestion(); q != nil {
				tipX := c.TopLeftX() + c.Width()/2 - s.camera.X
				tipY := c.TopLeftY() - s.camera.Y
				drawSpeechBubble(screen, s.fontFace, q.Question(), tipX, tipY)
			} else if result := c.InterviewResult(); result != nil {
				tipX := c.TopLeftX() + c.Width()/2 - s.camera.X
				tipY := c.TopLeftY() - s.camera.Y
				var msg string
				switch result.Outcome() {
				case entity.OutcomeOffer:
					msg = fmt.Sprintf("We are hiring you.\nYour salary will be $%d.\nDo you accept?", result.Salary())
				case entity.OutcomeRoundPassed:
					msg = "Great job!\nYou've passed this round."
				case entity.OutcomeRejection:
					msg = "..."
				}
				if len(msg) > 0 {
					drawSpeechBubble(screen, s.fontFace, msg, tipX, tipY)
				}
			}
		}
	}

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(s.player.TopLeftX()-s.camera.X, s.player.TopLeftY()-s.camera.Y)
	s.player.Draw(screen, op)

	for _, signal := range s.offscreenSignals {
		switch signal.signalType {
		case SignalTypeToken:
			vector.DrawFilledCircle(screen, float32(signal.x), float32(signal.y), 3, signal.color, true)
		case SignalTypeInterviewer:
			drawChevronSignal(screen, signal.x, signal.y, signal.angle, signal.color)
		}
	}

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

func drawChevronSignal(screen *ebiten.Image, x, y, angle float64, clr color.Color) {
	const size = 7.0
	const gap = 6.0
	const thickness = 2.0

	cos, sin := math.Cos(angle), math.Sin(angle)
	rotate := func(px, py float64) (float32, float32) {
		return float32(x + px*cos - py*sin), float32(y + px*sin + py*cos)
	}

	r, g, b, a := clr.RGBA()
	vclr := color.RGBA{R: uint8(r >> 8), G: uint8(g >> 8), B: uint8(b >> 8), A: uint8(a >> 8)}

	for i := range 2 {
		ox := -float64(i) * gap
		x1, y1 := rotate(ox-size, -size)
		x2, y2 := rotate(ox, 0)
		x3, y3 := rotate(ox-size, size)

		var p vector.Path
		p.MoveTo(x1, y1)
		p.LineTo(x2, y2)
		p.LineTo(x3, y3)
		vs, is := p.AppendVerticesAndIndicesForStroke(nil, nil, &vector.StrokeOptions{Width: thickness})
		for j := range vs {
			vs[j].SrcX, vs[j].SrcY = 1, 1
			vs[j].ColorR = float32(vclr.R) / 255
			vs[j].ColorG = float32(vclr.G) / 255
			vs[j].ColorB = float32(vclr.B) / 255
			vs[j].ColorA = float32(vclr.A) / 255
		}
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})
	}
}

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
