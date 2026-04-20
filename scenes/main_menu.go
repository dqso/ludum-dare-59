package scenes

import (
	"context"

	"github.com/dqso/ludum-dare-59/entity"
	"github.com/hajimehoshi/ebiten/v2"
)

type mainMenuScene struct {
	ctx  context.Context
	game entity.Game

	questions entity.QuestionsForInterview
}

func NewMainMenuScene(ctx context.Context, game entity.Game, questions entity.QuestionsForInterview) entity.Scene {
	return &mainMenuScene{
		ctx:       ctx,
		game:      game,
		questions: questions,
	}
}

func (s *mainMenuScene) Update() (entity.Scene, error) {
	// TODO
	return NewGenerateBattlefieldScene(s.ctx, s.game, s.questions), nil
}

func (s *mainMenuScene) Draw(screen *ebiten.Image) {}

func (s *mainMenuScene) Name() string { return "Main Menu" }
