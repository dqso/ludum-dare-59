package scenes

import (
	"context"
	"time"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/database"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/hajimehoshi/ebiten/v2"
)

type loadScene struct {
	ctx  context.Context
	game entity.Game

	status       uint8
	updateTicker *time.Ticker

	questions  entity.QuestionsForInterview
	firstNames entity.FirstNamesDatabase
}

func NewLoadScene(ctx context.Context, game entity.Game) entity.Scene {
	s := &loadScene{
		ctx:    ctx,
		game:   game,
		status: loadStartResources,
	}

	return s
}

const (
	_ uint8 = iota
	loadStartResources
	loadInProgressResources
	loadEndResources
)

func (s *loadScene) Update() (entity.Scene, error) {
	switch s.status {

	case loadStartResources:
		// Инициализировать загрузчики ресурсов
		s.updateTicker = time.NewTicker(time.Second / 60)
		s.status = loadInProgressResources

	case loadInProgressResources:
	loopInProgress:
		for {
			select {
			case <-s.updateTicker.C:
				break loopInProgress
			default:
			}
			if s.questions.Recruiter == nil {
				questions, err := database.NewQuestionsDatabase(assets.QuestionsRecruiterCSV)
				if err != nil {
					return NewErrorScene(s.ctx, s.game, err), nil
				}
				s.questions.Recruiter = questions
			} else if s.questions.Engineer == nil {
				questions, err := database.NewQuestionsDatabase(assets.QuestionsEngineerCSV)
				if err != nil {
					return NewErrorScene(s.ctx, s.game, err), nil
				}
				s.questions.Engineer = questions
			} else if s.firstNames == nil {
				firstNames, err := database.NewFirstNameDatabase(assets.FemaleFirstNamesTxt, assets.MaleFirstNamesTxt, assets.CompanyNamesTxt)
				if err != nil {
					return NewErrorScene(s.ctx, s.game, err), nil
				}
				s.firstNames = firstNames
			} else {
				s.status = loadEndResources
			}
		}

	case loadEndResources:
		// Почистить память

		return NewMainMenuScene(s.ctx, s.game, s.questions, s.firstNames), nil
	}
	return nil, nil
}

func (s *loadScene) Draw(_ *ebiten.Image) {}

func (s *loadScene) Name() string { return "Load" }
