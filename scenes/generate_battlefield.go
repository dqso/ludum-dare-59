package scenes

import (
	"context"
	"time"

	"github.com/dqso/ludum-dare-59/entity"
	"github.com/hajimehoshi/ebiten/v2"
)

type generateBattlefieldScene struct {
	ctx  context.Context
	game entity.Game

	status       uint8
	updateTicker *time.Ticker

	questions entity.QuestionsForInterview
}

func NewGenerateBattlefieldScene(ctx context.Context, game entity.Game, questions entity.QuestionsForInterview) entity.Scene {
	s := &generateBattlefieldScene{
		ctx:       ctx,
		game:      game,
		questions: questions,
		status:    generateBattlefieldStartResources,
	}

	return s
}

const (
	_ uint8 = iota
	generateBattlefieldStartResources
	generateBattlefieldInProgressResources
	generateBattlefieldEndResources
)

func (s *generateBattlefieldScene) Update() (entity.Scene, error) {
	switch s.status {

	case generateBattlefieldStartResources:
		// Инициализировать загрузчики ресурсов
		s.updateTicker = time.NewTicker(time.Second / 60)
		s.status = generateBattlefieldInProgressResources

	case generateBattlefieldInProgressResources:
	loopInProgress:
		for {
			select {
			case <-s.updateTicker.C:
				break loopInProgress
			default:
			}
			s.status = generateBattlefieldEndResources
		}

	case generateBattlefieldEndResources:
		// Почистить память

		return NewBattlefieldScene(s.ctx, s.game, s.questions), nil
	}
	return nil, nil
}

func (s *generateBattlefieldScene) Draw(screen *ebiten.Image) {}

func (s *generateBattlefieldScene) Name() string { return "Generate Battlefield" }
