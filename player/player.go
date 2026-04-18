package player

import (
	"bytes"
	"image"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/hajimehoshi/ebiten/v2"
)

type Player struct {
	x, y   float64
	w, h   float64
	speed  float64
	sprite *ebiten.Image
	// TODO пивот
}

const targetH = 45.0

func NewPlayer(x, y float64) (*Player, error) {
	img, _, err := image.Decode(bytes.NewReader(assets.MainCharPNG))
	if err != nil {
		return nil, err
	}
	src := ebiten.NewImageFromImage(img)

	scale := targetH / float64(src.Bounds().Dy())
	w := float64(src.Bounds().Dx()) * scale

	sprite := ebiten.NewImage(int(w), int(targetH))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	sprite.DrawImage(src, op)

	return &Player{
		x:      x,
		y:      y,
		w:      w,
		h:      targetH,
		speed:  5.0,
		sprite: sprite,
	}, nil
}

func (p Player) X() float64        { return p.x }
func (p Player) Y() float64        { return p.y }
func (p Player) TopLeftX() float64 { return p.x - p.w/2 }
func (p Player) TopLeftY() float64 { return p.y - p.h/2 }
func (p Player) Speed() float64    { return p.speed }

func (p *Player) Move(dx, dy float64) {
	p.x += dx
	p.y += dy
}

func (p Player) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	screen.DrawImage(p.sprite, op)
}
