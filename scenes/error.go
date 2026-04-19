package scenes

import (
	"log"

	"github.com/dqso/ludum-dare-59/entity"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type errorScene struct {
	game entity.Game
	err  error
}

func NewErrorScene(game entity.Game, err error) entity.Scene {
	log.Printf("created error scene with error: %v", err)
	return &errorScene{
		game: game,
		err:  err,
	}
}

func (s *errorScene) Update() (entity.Scene, error) {
	return nil, nil
}

func (s *errorScene) Draw(screen *ebiten.Image) {
	ebitenutil.DebugPrint(screen, s.err.Error())
}

func (s *errorScene) Name() string { return "Error" }
