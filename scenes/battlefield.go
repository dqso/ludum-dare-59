package scenes

import (
	"bytes"
	"context"
	"fmt"
	"image/color"
	"log"
	"maps"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/character"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/dqso/ludum-dare-59/player"
	"github.com/dqso/ludum-dare-59/token"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/shopspring/decimal"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

var globalAudioContext = audio.NewContext(44100)

type battlefieldScene struct {
	ctx        context.Context
	game       entity.Game
	questions  entity.QuestionsForInterview
	firstNames entity.FirstNamesDatabase

	stampatelloFaceto10 font.Face
	stampatelloFaceto14 font.Face
	cascadiaMono14      font.Face

	gameStart            time.Time
	nextTokenSpawn       time.Time
	nextInterviewerSpawn time.Time

	player       entity.Player
	camera       Camera
	characters   entity.CharacterList
	tokens       []entity.Token
	inventory    [9]entity.CollectedToken
	queueAnswers []AnswerWithPoint

	tokensInFocus    int
	offscreenSignals []offscreenSignal

	messages []Message
	expenses []*Expense

	options      entity.GameOptions
	minNumPoints int

	stats entity.Stats

	escMenuOpen bool

	musicTracks  [][]byte
	currentTrack int
	musicPlayer  *audio.Player
}

type AnswerWithPoint struct {
	answer entity.Answer
	x, y   float64
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

type Message struct {
	message   string
	expiredTo time.Time
}

type Expense struct {
	Expense     string
	Cost        func() decimal.Decimal
	ScheduledAt time.Time
	AddMonth    int
	AddDays     int
	AddDuration func() time.Duration
}

const (
	maxSpawnDistance = 1500.0

	tokenPoolSize           = 80
	tokenSpawnBatchSizeFrom = 1
	tokenSpawnBatchSizeTo   = 3
	tokenSpawnDelayFrom     = time.Second * 1
	tokenSpawnDelayTo       = time.Second * 2
	tokenLifetimeFrom       = time.Second * 50
	tokenLifetimeTo         = time.Second * 70

	interviewerRecruiterPoolSize = 5
	interviewerSpawnBatchSize    = 1
	interviewerSpawnDelayFrom    = time.Second * 60
	interviewerSpawnDelayTo      = time.Second * 90
	interviewerLifetimeFrom      = time.Minute * 3
	interviewerLifetimeTo        = time.Minute * 4

	startedMoneyFrom = 2000.00
	startedMoneyTo   = 3000.00
	startedSalary    = 0
	gameOverMoney    = -500.00
	winnerMoney      = 10000.00

	numRecruiterQuestions = 5
	numEngineerQuestions  = 6
	numFounderQuestions   = 3

	coeffRealTime time.Duration = 12 * 24 * 30 // 5 минут = 1 месяц
)

func (s *battlefieldScene) gameTimeNow() time.Time {
	return gameStartDate().Add(time.Since(s.gameStart) * coeffRealTime)
}

func gameStartDate() time.Time {
	return time.Date(2018, time.October, 6, 0, 0, 0, 0, time.UTC)
}

func moneyToString(money decimal.Decimal) string {
	return fmt.Sprintf("$%s", money.StringFixed(2))
}

func NewBattlefieldScene(ctx context.Context, game entity.Game, questions entity.QuestionsForInterview, firstNames entity.FirstNamesDatabase, options entity.GameOptions) entity.Scene {
	winW, winH := game.WindowSize()

	startedMoney := float64(startedMoneyFrom*100+rand.Int64N((startedMoneyTo-startedMoneyFrom)*100)) / 100
	player, err := player.NewPlayer(0, 0, decimal.NewFromFloat(startedMoney), decimal.NewFromFloat(startedSalary))
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	camera := Camera{
		X: player.X() - float64(winW)/2,
		Y: player.Y() - float64(winH)/2,
	}

	stampatelloFaceto, err := opentype.Parse(assets.StampatelloFacetoKernTTF)
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	stampatelloFaceto10, err := opentype.NewFace(stampatelloFaceto, &opentype.FaceOptions{
		Size:    float64(10),
		DPI:     96,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	stampatelloFaceto14, err := opentype.NewFace(stampatelloFaceto, &opentype.FaceOptions{
		Size:    float64(14),
		DPI:     96,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	cascadiaMono, err := opentype.Parse(assets.CascadiaMonoTTF)
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	cascadiaMono14, err := opentype.NewFace(cascadiaMono, &opentype.FaceOptions{
		Size:    float64(16),
		DPI:     96,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}

	expenses := []*Expense{
		{
			Expense:     "salary hack",
			Cost:        func() decimal.Decimal { return decimal.Zero },
			ScheduledAt: time.Date(gameStartDate().Year(), gameStartDate().Month()+1, 1, 12, 0, 0, 0, gameStartDate().Location()).AddDate(0, 0, -1),
			AddMonth:    1,
		},
		{
			Expense: "for groceries",
			Cost: func() decimal.Decimal {
				return decimal.NewFromFloat(200 + float64(rand.IntN(10000))/100)
			},
			ScheduledAt: gameStartDate().Add(time.Hour * 24 * 7),
			AddDays:     7,
			AddDuration: func() time.Duration { return randDuration(-time.Hour*3, time.Hour*3) },
		},
	}
	if !options.DecreaseDifficulty {
		expenses = append(expenses, &Expense{
			Expense:     "for rent",
			Cost:        func() decimal.Decimal { return decimal.NewFromFloat(700) },
			ScheduledAt: time.Date(gameStartDate().Year(), gameStartDate().Month()+1, 1, 12, 0, 0, 0, gameStartDate().Location()),
			AddMonth:    1,
		})
	}

	musicTracks := entity.GetMusicPlaylist()
	rand.Shuffle(len(musicTracks), func(i, j int) {
		musicTracks[i], musicTracks[j] = musicTracks[j], musicTracks[i]
	})

	minNumPoints := 5
	if options.DecreaseDifficulty {
		minNumPoints = 2
	}

	now := time.Now()
	return &battlefieldScene{
		ctx:        ctx,
		game:       game,
		questions:  questions,
		firstNames: firstNames,

		stampatelloFaceto10: stampatelloFaceto10,
		stampatelloFaceto14: stampatelloFaceto14,
		cascadiaMono14:      cascadiaMono14,

		gameStart:            now,
		nextTokenSpawn:       now,
		nextInterviewerSpawn: now.Add(time.Second * 3),

		player:     player,
		camera:     camera,
		characters: character.NewList(),
		tokens:     make([]entity.Token, 0),

		messages: make([]Message, 0),
		expenses: expenses,

		options:      options,
		minNumPoints: minNumPoints,
		musicTracks:  musicTracks,
		currentTrack: -1,
	}
}

func (s *battlefieldScene) playNotification() {
	decoded, err := mp3.DecodeWithSampleRate(globalAudioContext.SampleRate(), bytes.NewReader(assets.NotificationMP3))
	if err != nil {
		log.Printf("notification decode error: %v", err)
		return
	}
	p, err := globalAudioContext.NewPlayer(decoded)
	if err != nil {
		log.Printf("notification player error: %v", err)
		return
	}
	p.SetVolume(s.options.SoundVolume)
	p.Play()
}

func (s *battlefieldScene) playTrack(idx int) {
	if s.musicPlayer != nil {
		s.musicPlayer.Pause()
		if err := s.musicPlayer.Close(); err != nil {
			log.Printf("musicPlayer close error: %v", err)
		}
		s.musicPlayer = nil
	}
	decoded, err := mp3.DecodeWithSampleRate(globalAudioContext.SampleRate(), bytes.NewReader(s.musicTracks[idx]))
	if err != nil {
		log.Printf("music decode error: %v", err)
		return
	}
	p, err := globalAudioContext.NewPlayer(decoded)
	if err != nil {
		log.Printf("music player error: %v", err)
		return
	}
	p.SetVolume(s.options.MusicVolume)
	p.Play()
	s.musicPlayer = p
}

func (s *battlefieldScene) Update() (entity.Scene, error) {
	lose := s.player.Money().LessThan(decimal.NewFromFloat(gameOverMoney))
	win := s.player.Money().GreaterThan(decimal.NewFromFloat(winnerMoney))
	if lose || win {
		s.stats.DurationInGame = s.gameTimeNow().Sub(gameStartDate())
		s.stats.RealTimePlayed = time.Since(s.gameStart)
		s.stats.CurrentBalance = s.player.Money()
		s.stats.MonthlySalary = s.player.SalaryPerMonth()
		_type := typeGameOverLose
		if win {
			_type = typeGameOverWin
		}
		return NewGameOverScene(s.ctx, _type, s.game, s.stampatelloFaceto10, s.questions, s.firstNames, s.stats, s.options), nil
	}
	if s.options.MusicVolume > 0 && (s.musicPlayer == nil || !s.musicPlayer.IsPlaying()) {
		s.currentTrack = (s.currentTrack + 1) % len(s.musicTracks)
		s.playTrack(s.currentTrack)
	}

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
				if founder.PlayerPoints() >= s.minNumPoints {
					founder.SetInterviewResult(character.NewRoundPassed())
				} else {
					founder.SetInterviewResult(character.NewRejection())
					break idleFor
				}
			}
			recruiter.SetInterviewResult(character.NewOffer(decimal.NewFromFloat(float64(2000 + rand.IntN(5)*100))))
			s.stats.OffersReceived++
			duration := randDuration(interviewerLifetimeFrom, interviewerLifetimeTo)
			recruiter.UpdateDeadline(duration)
			if s.options.TechInterview {
				engineer.UpdateDeadline(duration)
			}
			founder.UpdateDeadline(duration)
		} else if engineer != nil && character.FilterIdle(engineer) && s.options.TechInterview {
			if engineer.GetQuestion() == nil {
				if engineer.PlayerPoints() >= s.minNumPoints {
					engineer.SetInterviewResult(character.NewRoundPassed())
				} else {
					engineer.SetInterviewResult(character.NewRejection())
					continue idleFor
				}
			}
			duration := randDuration(interviewerLifetimeFrom, interviewerLifetimeTo)
			questions, answers := s.questions.Recruiter.GetRandomQuestions(3) // TODO fix
			s.queueAnswers = append(getSomeAnswers(answers, spawnX, spawnY), s.queueAnswers...)
			newChar, err := character.NewCharacter(s.firstNames.GetRandomMaleFirstName(),
				entity.CharacterRoleFounder,
				spawnX, spawnY,
				questions,
				duration,
				c.Company(),
			)
			if err != nil {
				return NewErrorScene(s.ctx, s.game, err), nil
			}
			s.characters.Add(newChar)
			recruiter.UpdateDeadline(duration)
			engineer.UpdateDeadline(duration)
		} else if recruiter != nil && character.FilterIdle(recruiter) {
			if recruiter.GetQuestion() == nil {
				if recruiter.PlayerPoints() >= s.minNumPoints {
					recruiter.SetInterviewResult(character.NewRoundPassed())
				} else {
					recruiter.SetInterviewResult(character.NewRejection())
					continue idleFor
				}
			}
			duration := randDuration(interviewerLifetimeFrom, interviewerLifetimeTo)
			var (
				nextRole  entity.CharacterRole
				questions []entity.Question
				answers   entity.AnswersGrouped
			)
			if !s.options.TechInterview {
				nextRole = entity.CharacterRoleFounder
				questions, answers = s.questions.Recruiter.GetRandomQuestions(numFounderQuestions) // TODO
			} else {
				nextRole = entity.CharacterRoleEngineer
				questions, answers = s.questions.Engineer.GetRandomQuestions(numEngineerQuestions)
			}
			s.queueAnswers = append(getSomeAnswers(answers, spawnX, spawnY), s.queueAnswers...)
			newChar, err := character.NewCharacter(s.firstNames.GetRandomMaleFirstName(),
				nextRole,
				spawnX, spawnY,
				questions,
				duration,
				c.Company(),
			)
			if err != nil {
				return NewErrorScene(s.ctx, s.game, err), nil
			}
			s.characters.Add(newChar)
			recruiter.UpdateDeadline(duration)
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
		for i := len(s.queueAnswers) - 1; i >= 0 && tokensToSpawn > 0; i-- {
			tokensToSpawn--
			var a AnswerWithPoint
			a, s.queueAnswers = s.queueAnswers[i], s.queueAnswers[:i]
			spawnAngle := rand.Float64() * 2 * math.Pi
			spawnR := 200 + rand.Float64()*(1000-200)
			t := token.NewToken(s.stampatelloFaceto14, a.answer,
				a.x+spawnR*math.Cos(spawnAngle),
				a.y+spawnR*math.Sin(spawnAngle),
				randDuration(tokenLifetimeFrom, tokenLifetimeTo),
			)
			s.tokens = append(s.tokens, t)
		}
		rnd := []entity.QuestionsDatabase{
			s.questions.Recruiter,
		}
		if s.options.TechInterview {
			rnd = append(rnd, s.questions.Engineer)
		}
		// TODO добавить ещё ответы овнера
		rand.Shuffle(len(rnd), func(i, j int) {
			rnd[i], rnd[j] = rnd[j], rnd[i]
		})
		for _, q := range rnd {
			for _, a := range q.GetRandomAnswers(tokensToSpawn / 2) {
				spawnAngle := rand.Float64() * 2 * math.Pi
				spawnR := 200 + rand.Float64()*(1000-200)
				t := token.NewToken(s.stampatelloFaceto14, a,
					s.player.X()+spawnR*math.Cos(spawnAngle),
					s.player.Y()+spawnR*math.Sin(spawnAngle),
					randDuration(time.Second*20, time.Second*30),
				)
				s.tokens = append(s.tokens, t)
			}
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
			company := s.firstNames.GetRandomCompanyName()
			questions, answers := s.questions.Recruiter.GetRandomQuestions(numRecruiterQuestions)
			x := float64(rand.IntN(maxSpawnDistance*2) - maxSpawnDistance)
			y := float64(rand.IntN(maxSpawnDistance*2) - maxSpawnDistance)
			s.queueAnswers = append(getSomeAnswers(answers, x, y), s.queueAnswers...)
			firstName := s.firstNames.GetRandomFemaleFirstName()
			c, err := character.NewCharacter(firstName, entity.CharacterRoleRecruiter,
				x, y,
				questions,
				randDuration(interviewerLifetimeFrom, interviewerLifetimeTo),
				company,
			)
			if err != nil {
				return NewErrorScene(s.ctx, s.game, err), nil
			}
			s.characters.Add(c)
			s.playNotification()
			msg := fmt.Sprintf(
				"Hey, everyone! My name is %s.\nI represent %s.\nWe're looking for a developer to join us.",
				firstName, company,
			)
			s.messages = append(s.messages, Message{
				message:   msg,
				expiredTo: time.Now().Add(time.Second * 8),
			})
		}
	}

	if ebiten.IsKeyPressed(ebiten.KeyF) && ebiten.IsKeyPressed(ebiten.KeyShift) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.escMenuOpen = !s.escMenuOpen
	}
	if s.escMenuOpen && inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		winW, winH := s.game.WindowSize()
		mx, my := ebiten.CursorPosition()
		btnX, btn1Y, btn2Y := escMenuButtonPositions(winW, winH)
		switch {
		case escMenuInRect(mx, my, btnX, btn1Y):
			s.musicPlayer.Pause()
			if err := s.musicPlayer.Close(); err != nil {
				log.Printf("failed to close music player: %v", err)
			}
			s.musicPlayer = nil
			return NewMainMenuScene(s.ctx, s.game, s.questions, s.firstNames, s.options), nil
		case escMenuInRect(mx, my, btnX, btn2Y):
			s.escMenuOpen = false
		}
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

			if nearestCharacter != nil && nearestCharacter.GetQuestion() != nil {
				answer := t.Answer()
				nearestCharacter.AnswerTheQuestion(s.questions.Choose(nearestCharacter), answer)
				s.stats.QuestionsAnswered++
			} else if nearestCharacter != nil && (nearestCharacter.InterviewResult() != nil && nearestCharacter.InterviewResult().Outcome() == entity.OutcomeOffer) {
				answer := t.Answer()
				switch answer.Category() {
				case entity.AnswerCategoryBinaryYes:
					salary := nearestCharacter.InterviewResult().Salary()
					for c := range s.characters.DeleteFunc(func(c entity.Character) bool {
						return c.Company() == nearestCharacter.Company()
					}) {
						log.Printf("Character %s from company %q has been removed.", c.Role(), c.Company())
					}
					s.player.AddSalaryPerMonth(salary)
					s.stats.OffersAccepted++
				case entity.AnswerCategoryBinaryNo:
					for c := range s.characters.DeleteFunc(func(c entity.Character) bool {
						return c.Company() == nearestCharacter.Company()
					}) {
						log.Printf("Character %s from company %q has been removed.", c.Role(), c.Company())
					}
					s.stats.OffersRejected++
				default:
					s.tokens = append(s.tokens, t)
				}
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
		const hotbarH = 80
		scaleX := (winW64/2 - margin) / math.Abs(dx)
		var scaleY float64
		if dy > 0 {
			scaleY = (winH64 - hotbarH - margin - cy) / dy
		} else {
			scaleY = (cy - margin) / math.Abs(dy)
		}
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

	for i := len(s.messages) - 1; i >= 0; i-- {
		if time.Since(s.messages[i].expiredTo) > 0 {
			s.messages = append(s.messages[:i], s.messages[i+1:]...)
		}
	}

	for _, e := range s.expenses {
		if s.gameTimeNow().Sub(e.ScheduledAt) > 0 {
			e.ScheduledAt = e.ScheduledAt.AddDate(0, e.AddMonth, e.AddDays)
			if e.AddDuration != nil {
				e.ScheduledAt = e.ScheduledAt.Add(e.AddDuration())
			}

			var msg string
			if e.Expense == "salary hack" {
				salary := s.player.SalaryPerMonth()
				if salary.LessThanOrEqual(decimal.Zero) {
					continue
				}
				msg = fmt.Sprintf("You earned %s this month.", moneyToString(salary))
				s.player.AddMoney(salary)
				s.stats.Earned.Add(salary)
			} else {
				cost := e.Cost()
				msg = fmt.Sprintf("You paid %s %s.", moneyToString(cost), e.Expense)
				s.player.AddMoney(cost.Neg())
				s.stats.Spent.Add(cost)
			}

			if len(msg) > 0 {
				s.messages = append(s.messages, Message{
					message:   msg,
					expiredTo: time.Now().Add(time.Second * 3),
				})
			}
		}
	}

	return nil, nil
}

func (s *battlefieldScene) Draw(screen *ebiten.Image) {
	drawBackground(screen, s.camera.X, s.camera.Y)
	winW, winH := s.game.WindowSize()

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
				drawSpeechBubble(screen, s.stampatelloFaceto10, s.stampatelloFaceto14, titleOfCharacter(c), q.Question(), tipX, tipY)
			} else if result := c.InterviewResult(); result != nil {
				tipX := c.TopLeftX() + c.Width()/2 - s.camera.X
				tipY := c.TopLeftY() - s.camera.Y
				var msg string
				switch result.Outcome() {
				case entity.OutcomeOffer:
					msg = fmt.Sprintf("We are hiring you.\nYour salary will be %s.\nDo you accept?", moneyToString(result.Salary()))
				case entity.OutcomeRoundPassed:
					msg = "Great job!\nYou've passed this round."
				case entity.OutcomeRejection:
					msg = "..."
				}
				if len(msg) > 0 {
					drawSpeechBubble(screen, s.stampatelloFaceto10, s.stampatelloFaceto14, titleOfCharacter(c), msg, tipX, tipY)
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

	drawInventory(screen, s.stampatelloFaceto10, s.inventory[:], winW, winH)

	timeStr := s.gameTimeNow().Format("Mon,_2 Jan 2006 15:04")
	timeBounds := text.BoundString(s.cascadiaMono14, timeStr)
	timeOp := &ebiten.DrawImageOptions{}
	timeOp.GeoM.Translate(float64(winW)-15-float64(timeBounds.Dx())-float64(timeBounds.Min.X), 15-float64(timeBounds.Min.Y))
	timeOp.ColorM.ScaleWithColor(color.NRGBA{R: 200, G: 210, B: 255, A: 220})
	text.DrawWithOptions(screen, timeStr, s.cascadiaMono14, timeOp)

	moneyStr := moneyToString(s.player.Money())
	moneyBounds := text.BoundString(s.cascadiaMono14, moneyStr)
	moneyOp := &ebiten.DrawImageOptions{}
	moneyOp.GeoM.Translate(float64(winW)-15-float64(moneyBounds.Dx())-float64(moneyBounds.Min.X), 15*2+float64(timeBounds.Dy())-float64(moneyBounds.Min.Y))
	moneyOp.ColorM.ScaleWithColor(color.NRGBA{R: 200, G: 210, B: 255, A: 220})
	text.DrawWithOptions(screen, moneyStr, s.cascadiaMono14, moneyOp)

	salaryStr := moneyToString(s.player.SalaryPerMonth()) + "/month"
	salaryBounds := text.BoundString(s.cascadiaMono14, salaryStr)
	salaryOp := &ebiten.DrawImageOptions{}
	salaryOp.GeoM.Translate(float64(winW)-15-float64(salaryBounds.Dx())-float64(salaryBounds.Min.X), 15*3+float64(timeBounds.Dy())+float64(moneyBounds.Dy())-float64(salaryBounds.Min.Y))
	salaryOp.ColorM.ScaleWithColor(color.NRGBA{R: 200, G: 210, B: 255, A: 220})
	text.DrawWithOptions(screen, salaryStr, s.cascadiaMono14, salaryOp)

	const hotbarH = 80 + 10 // barH + sideMargin
	drawToasts(screen, s.stampatelloFaceto10, s.messages, float64(winW), float64(winH)-hotbarH)

	if s.options.ShowFPS {
		fpsStr := fmt.Sprintf("%d", int(ebiten.ActualFPS()))
		fpsBounds := text.BoundString(s.cascadiaMono14, fpsStr)
		fpsOp := &ebiten.DrawImageOptions{}
		fpsOp.GeoM.Translate(8-float64(fpsBounds.Min.X), 8-float64(fpsBounds.Min.Y))
		fpsOp.ColorM.ScaleWithColor(color.NRGBA{R: 80, G: 80, B: 80, A: 160})
		text.DrawWithOptions(screen, fpsStr, s.cascadiaMono14, fpsOp)
	}

	if s.escMenuOpen {
		s.drawEscMenu(screen, winW, winH)
	}
}

func (s *battlefieldScene) Name() string { return "Battlefield" }

var circuitCache = map[[2]int]int{} // col,row -> тип

func cellType(col, row int) int {
	key := [2]int{col, row}
	if t, ok := circuitCache[key]; ok {
		return t
	}
	// детерминированный рандом по координатам
	h := col*73856093 ^ row*19349663
	t := ((h % 5) + 5) % 5 // 0=пусто 1=горизонталь 2=вертикаль 3=угол 4=точка
	circuitCache[key] = t
	return t
}

func drawBackground(screen *ebiten.Image, cameraX, cameraY float64) {
	const S = 40.0
	W := float64(screen.Bounds().Dx())
	H := float64(screen.Bounds().Dy())

	ox := math.Mod(-cameraX, S)
	if ox < 0 {
		ox += S
	}
	oy := math.Mod(-cameraY, S)
	if oy < 0 {
		oy += S
	}

	clr := color.RGBA{15, 35, 15, 50}

	colStart := int(math.Floor(cameraX / S))
	rowStart := int(math.Floor(cameraY / S))

	for ci := -1; float64(ci)*S+ox < W+S; ci++ {
		for ri := -1; float64(ri)*S+oy < H+S; ri++ {
			x := float32(float64(ci)*S + ox)
			y := float32(float64(ri)*S + oy)
			mx, my := x+S/2, y+S/2

			switch cellType(colStart+ci, rowStart+ri) {
			case 1: // горизонталь
				vector.StrokeLine(screen, x, my, x+S, my, 1, clr, false)
			case 2: // вертикаль
				vector.StrokeLine(screen, mx, y, mx, y+S, 1, clr, false)
			case 3: // угол
				vector.StrokeLine(screen, x, my, mx, my, 1, clr, false)
				vector.StrokeLine(screen, mx, my, mx, y, 1, clr, false)
			case 4: // точка
				vector.DrawFilledCircle(screen, mx, my, 3, clr, false)
			}
		}
	}
}

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

func drawSpeechBubble(screen *ebiten.Image, face10, face14 font.Face, name, msg string, tipX, tipY float64) {
	const pad = 10.0
	const r = 8.0
	const arrowH = 14.0
	const arrowHalfW = 8.0

	boundsName := text.BoundString(face10, name)
	boundsMsg := text.BoundString(face14, msg)
	w := math.Max(float64(boundsName.Dx()), float64(boundsMsg.Dx())) + pad*2
	h := float64(boundsName.Dy()) + float64(boundsMsg.Dy()) + pad*3

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
	op.ColorM.ScaleWithColor(colornames.Darkred)
	op.GeoM.Translate(x+pad-float64(boundsName.Min.X), y+pad-float64(boundsName.Min.Y))
	text.DrawWithOptions(screen, name, face10, op)

	op = &ebiten.DrawImageOptions{}
	op.ColorM.ScaleWithColor(colornames.Black)
	op.GeoM.Translate(x+pad-float64(boundsMsg.Min.X), y+pad+float64(boundsName.Dy())+pad-float64(boundsMsg.Min.Y))
	text.DrawWithOptions(screen, msg, face14, op)
}

func drawToasts(screen *ebiten.Image, face font.Face, messages []Message, winW, winH float64) {
	const pad = 10.0
	const margin = 10.0
	const maxWidth = 300.0
	const maxHeight = 200.0

	y := winH - margin
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i].message
		bounds := text.BoundString(face, msg)
		w := math.Min(float64(bounds.Dx())+pad*2, maxWidth)
		h := math.Min(float64(bounds.Dy())+pad*2, maxHeight)
		x := winW - margin*2 - w
		y -= h

		var bg vector.Path
		const r = 5.0
		bg.MoveTo(float32(x+r), float32(y))
		bg.LineTo(float32(x+w-r), float32(y))
		bg.ArcTo(float32(x+w), float32(y), float32(x+w), float32(y+r), r)
		bg.LineTo(float32(x+w), float32(y+h-r))
		bg.ArcTo(float32(x+w), float32(y+h), float32(x+w-r), float32(y+h), r)
		bg.LineTo(float32(x+r), float32(y+h))
		bg.ArcTo(float32(x), float32(y+h), float32(x), float32(y+h-r), r)
		bg.LineTo(float32(x), float32(y+r))
		bg.ArcTo(float32(x), float32(y), float32(x+r), float32(y), r)
		bg.Close()
		vs, is := bg.AppendVerticesAndIndicesForFilling(nil, nil)
		for j := range vs {
			vs[j].SrcX, vs[j].SrcY = 1, 1
			vs[j].ColorR, vs[j].ColorG, vs[j].ColorB, vs[j].ColorA = 0.98, 0.98, 0.9, 0.95
		}
		screen.DrawTriangles(vs, is, whiteImage, &ebiten.DrawTrianglesOptions{AntiAlias: true})

		op := &ebiten.DrawImageOptions{}
		op.ColorM.ScaleWithColor(colornames.Black)
		op.GeoM.Translate(x+pad-float64(bounds.Min.X), y+h/2-float64(bounds.Dy())/2-float64(bounds.Min.Y))
		text.DrawWithOptions(screen, msg, face, op)

		y -= margin
	}
}

func drawInventory(screen *ebiten.Image, face font.Face, slots []entity.CollectedToken, winW, winH int) {
	const (
		barH     float32 = 80.0
		padY     float32 = 8.0
		innerPad float32 = 4.0
	)

	const sideMargin float32 = 10.0

	n := len(slots)
	barW := float32(winW) - sideMargin*2
	slotW := barW / float32(n)
	barY := float32(winH) - barH

	vector.DrawFilledRect(screen, sideMargin, barY, barW, barH, color.NRGBA{10, 10, 25, 210}, true)
	vector.StrokeRect(screen, sideMargin, barY, barW, barH, 1, color.NRGBA{60, 70, 110, 180}, true)

	for i, ct := range slots {
		sx := sideMargin + float32(i)*slotW
		sy := barY

		filled := ct != nil
		bgClr := color.NRGBA{20, 20, 45, 0}
		borderClr := color.NRGBA{55, 60, 100, 160}
		if filled {
			bgClr = color.NRGBA{35, 40, 80, 180}
			borderClr = color.NRGBA{130, 150, 230, 255}
		}
		vector.DrawFilledRect(screen, sx+innerPad, sy+innerPad, slotW-innerPad*2, barH-innerPad*2, bgClr, true)
		vector.StrokeRect(screen, sx+innerPad, sy+innerPad, slotW-innerPad*2, barH-innerPad*2, 1.5, borderClr, true)

		slotInnerW := slotW - innerPad*2
		slotInnerH := barH - innerPad*2 - padY

		if filled {
			scale := float32(1.0)
			scaleX := slotInnerW / float32(ct.Width())
			scaleY := slotInnerH / float32(ct.Height())
			if scaleX < scale {
				scale = scaleX
			}
			if scaleY < scale {
				scale = scaleY
			}
			scaledW := float32(ct.Width()) * scale
			scaledH := float32(ct.Height()) * scale
			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(float64(scale), float64(scale))
			op.GeoM.Translate(
				float64(sx+innerPad)+(float64(slotInnerW)-float64(scaledW))/2,
				float64(sy+innerPad)+(float64(slotInnerH)-float64(scaledH))/2,
			)
			ct.Draw(screen, op)
		}

		label := strconv.Itoa(i + 1)
		bounds := text.BoundString(face, label)
		lop := &ebiten.DrawImageOptions{}
		lop.GeoM.Translate(float64(sx+innerPad)+3-float64(bounds.Min.X), float64(sy+barH-innerPad)-float64(bounds.Dy())-2-float64(bounds.Min.Y))
		if filled {
			lop.ColorM.ScaleWithColor(color.RGBA{200, 210, 255, 255})
		} else {
			lop.ColorM.ScaleWithColor(color.RGBA{70, 75, 110, 180})
		}
		text.DrawWithOptions(screen, label, face, lop)
	}
}

func titleOfCharacter(c interface {
	Name() string
	Role() entity.CharacterRole
	Company() string
}) string {
	return fmt.Sprintf("%s | %s at %s", c.Name(), strings.Title(c.Role().String()), c.Company())
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

const (
	escMenuW      = 260.0
	escMenuH      = 150.0
	escMenuBtnW   = 200.0
	escMenuBtnH   = 42.0
	escMenuBtnGap = 12.0
	escMenuPadTop = 18.0
)

func escMenuButtonPositions(winW, winH int) (btnX, btn1Y, btn2Y float64) {
	panelX := (float64(winW) - escMenuW) / 2
	panelY := (float64(winH) - escMenuH) / 2
	btnX = panelX + (escMenuW-escMenuBtnW)/2
	btn1Y = panelY + escMenuPadTop
	btn2Y = btn1Y + escMenuBtnH + escMenuBtnGap
	return
}

func escMenuInRect(mx, my int, bx, by float64) bool {
	return float64(mx) >= bx && float64(mx) <= bx+escMenuBtnW &&
		float64(my) >= by && float64(my) <= by+escMenuBtnH
}

func (s *battlefieldScene) drawEscMenu(screen *ebiten.Image, winW, winH int) {
	vector.DrawFilledRect(screen, 0, 0, float32(winW), float32(winH), color.NRGBA{0, 0, 0, 140}, false)

	panelX := (float64(winW) - escMenuW) / 2
	panelY := (float64(winH) - escMenuH) / 2
	vector.DrawFilledRect(screen, float32(panelX), float32(panelY), escMenuW, escMenuH, color.NRGBA{15, 15, 30, 245}, false)
	vector.StrokeRect(screen, float32(panelX), float32(panelY), escMenuW, escMenuH, 1, color.NRGBA{80, 90, 140, 255}, false)

	mx, my := ebiten.CursorPosition()
	btnX, btn1Y, btn2Y := escMenuButtonPositions(winW, winH)

	drawBtn := func(label string, by float64) {
		hover := escMenuInRect(mx, my, btnX, by)
		bg := color.NRGBA{40, 40, 65, 255}
		border := color.NRGBA{100, 110, 160, 255}
		if hover {
			bg = color.NRGBA{65, 65, 100, 255}
			border = color.NRGBA{160, 170, 220, 255}
		}
		vector.DrawFilledRect(screen, float32(btnX), float32(by), escMenuBtnW, escMenuBtnH, bg, false)
		vector.StrokeRect(screen, float32(btnX), float32(by), escMenuBtnW, escMenuBtnH, 1, border, false)

		bounds := text.BoundString(s.stampatelloFaceto14, label)
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(
			btnX+(escMenuBtnW-float64(bounds.Dx()))/2-float64(bounds.Min.X),
			by+(escMenuBtnH-float64(bounds.Dy()))/2-float64(bounds.Min.Y),
		)
		op.ColorM.ScaleWithColor(color.RGBA{200, 210, 255, 255})
		text.DrawWithOptions(screen, label, s.stampatelloFaceto14, op)
	}

	drawBtn("Main Menu", btn1Y)
	drawBtn("Resume", btn2Y)
}

func getSomeAnswers(grouped entity.AnswersGrouped, x, y float64) []AnswerWithPoint {
	out := make([]AnswerWithPoint, 0)

	const maximum = 2

	get := func(m map[entity.Answer]struct{}) {
		answers := make([]AnswerWithPoint, 0)
		for k := range maps.Keys(m) {
			answers = append(answers, AnswerWithPoint{
				answer: k,
				x:      x,
				y:      y,
			})
		}
		rand.Shuffle(len(answers), func(i, j int) {
			answers[i], answers[j] = answers[j], answers[i]
		})
		if len(answers) > maximum {
			answers = answers[:maximum]
		}
		out = append(out, answers...)
	}

	get(grouped.Positive)
	get(grouped.Neutral)
	get(grouped.Negative)

	return out
}
