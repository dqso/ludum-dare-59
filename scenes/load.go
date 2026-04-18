package scenes

import (
	"time"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/database"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/hajimehoshi/ebiten/v2"
)

type loadScene struct {
	game entity.Game

	status       uint8
	updateTicker *time.Ticker

	questions entity.QuestionsDatabase
}

func NewLoadScene(game entity.Game) entity.Scene {
	s := &loadScene{
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
			if s.questions == nil {
				questions, err := database.NewQuestionsDatabase(assets.AnswersCSV)
				if err != nil {
					return NewErrorScene(s.game, err), nil
				}
				s.questions = questions
			} else {
				s.status = loadEndResources
			}
		}

	case loadEndResources:
		// Почистить память

		return NewMainMenuScene(s.game, s.questions), nil
	}
	return nil, nil
}

func (s *loadScene) Draw(screen *ebiten.Image) {}

func (s *loadScene) Name() string { return "Load" }
