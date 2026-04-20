package game

import (
	"context"
	"fmt"

	"github.com/dqso/ludum-dare-59/entity"
	"github.com/dqso/ludum-dare-59/scenes"
	"github.com/hajimehoshi/ebiten/v2"
)

type game struct {
	ctx   context.Context
	scene entity.Scene

	width, height int
}

func New(ctx context.Context) entity.Game {
	g := &game{
		ctx: ctx,
	}

	g.scene = scenes.NewLoadScene(ctx, g)

	return g
}

func (g *game) Update() error {
	select {
	case <-g.ctx.Done():
		return fmt.Errorf("game ended by OS signal")
	default:
	}

	if g.scene == nil {
		g.scene = scenes.NewErrorScene(g.ctx, g, fmt.Errorf("scene is nil"))
		return nil
	}
	newScene, err := g.scene.Update()
	if err != nil {
		g.scene = scenes.NewErrorScene(g.ctx, g, err)
		return nil
	} else {
		if newScene != nil {
			g.scene = newScene
		}
	}
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	g.scene.Draw(screen)
}

func (g *game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	g.width, g.height = outsideWidth, outsideHeight
	return g.width, g.height
}

func (g *game) WindowSize() (int, int) { return g.width, g.height }
