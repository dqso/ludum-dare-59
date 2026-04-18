package token

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

type Token struct {
	x, y   float64
	w, h   float64
	text   string
	sprite *ebiten.Image
}

func NewToken(fontFace font.Face, str string, x, y float64, clr color.Color) Token {
	rect := text.BoundString(fontFace, str)
	const padding = 5
	sprite := ebiten.NewImage(rect.Dx()+padding*2, rect.Dy()+padding*2)
	r32, g32, b32, a32 := clr.RGBA()
	r, g, b, a := uint8(r32>>8), uint8(g32>>8), uint8(b32>>8), uint8(a32>>8)
	//sprite.Fill(color.RGBA{
	//	R: r + (255-r)/2,
	//	G: g + (255-g)/2,
	//	B: b + (255-b)/2,
	//	A: a,
	//})
	_ = a
	luma := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b) // яркость по формуле luma
	var shift float64
	if luma > 128 {
		shift = -150 // светлый текст → тёмный
	} else {
		shift = +150 // тёмный текст → светлый фон
	}
	clamp := func(v float64) uint8 {
		if v < 0 {
			return 0
		}
		if v > 255 {
			return 255
		}
		return uint8(v)
	}
	sprite.Fill(color.RGBA{
		R: clamp(float64(r) + shift),
		G: clamp(float64(g) + shift),
		B: clamp(float64(b) + shift),
		A: 200,
	})

	vector.StrokeRect(sprite, 0, 0, float32(rect.Dx()-1)+padding*2, float32(rect.Dy()-1)+padding*2, 1, clr, false)
	var opts ebiten.DrawImageOptions
	opts.ColorM.ScaleWithColor(clr)
	opts.GeoM.Translate(float64(-rect.Min.X)+padding, float64(-rect.Min.Y)+padding)
	text.DrawWithOptions(sprite, str, fontFace, &opts)

	return Token{
		x:      x,
		y:      y,
		w:      float64(rect.Dx()) + padding*2,
		h:      float64(rect.Dy()) + padding*2,
		text:   str,
		sprite: sprite,
	}
}

func (t Token) X() float64        { return t.x }
func (t Token) Y() float64        { return t.y }
func (t Token) TopLeftX() float64 { return t.x - t.w/2 }
func (t Token) TopLeftY() float64 { return t.y - t.h/2 }

func (t Token) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	screen.DrawImage(t.sprite, op)
}
