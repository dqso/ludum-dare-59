package token

import (
	"image/color"

	"github.com/dqso/ludum-dare-59/entity"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

const padding = 5

type Token struct {
	x, y        float64
	w, h        float64
	sprite      *ebiten.Image
	lastFocused bool
	focused     bool
	fontFace    font.Face
	answer      entity.Answer
}

func NewToken(fontFace font.Face, answer entity.Answer, x, y float64) *Token {
	t := &Token{
		x:        x,
		y:        y,
		fontFace: fontFace,
		answer:   answer,
	}
	t.rebuildSprite()
	t.w = float64(t.sprite.Bounds().Dx())
	t.h = float64(t.sprite.Bounds().Dy())

	return t
}

func (t Token) X() float64                { return t.x }
func (t Token) Y() float64                { return t.y }
func (t Token) Width() float64            { return t.w }
func (t Token) Height() float64           { return t.h }
func (t Token) TopLeftX() float64         { return t.x - t.w/2 }
func (t Token) TopLeftY() float64         { return t.y - t.h/2 }
func (t *Token) SetFocus(focus bool)      { t.focused = focus }
func (t Token) IsFocused() bool           { return t.focused }
func (t *Token) SetPosition(x, y float64) { t.x = x; t.y = y }
func (t Token) Answer() entity.Answer     { return t.answer }

func (t *Token) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	if t.lastFocused != t.focused {
		t.rebuildSprite()
		t.lastFocused = t.focused
	}
	screen.DrawImage(t.sprite, op)
}

func (t *Token) rebuildSprite() {
	rect := text.BoundString(t.fontFace, t.answer.Answer())
	sprite := ebiten.NewImage(rect.Dx()+padding*2, rect.Dy()+padding*2)
	r32, g32, b32, a32 := entity.AnswerCategoryToColor(t.answer.Category()).RGBA()
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
	if t.focused {
		shift /= 2
	}
	sprite.Fill(color.RGBA{
		R: clamp(float64(r) + shift),
		G: clamp(float64(g) + shift),
		B: clamp(float64(b) + shift),
		A: 200,
	})

	vector.StrokeRect(sprite, 0, 0, float32(rect.Dx()-1)+padding*2, float32(rect.Dy()-1)+padding*2, 1, entity.AnswerCategoryToColor(t.answer.Category()), false)
	var opts ebiten.DrawImageOptions
	opts.ColorM.ScaleWithColor(entity.AnswerCategoryToColor(t.answer.Category()))
	opts.GeoM.Translate(float64(-rect.Min.X)+padding, float64(-rect.Min.Y)+padding)
	text.DrawWithOptions(sprite, t.answer.Answer(), t.fontFace, &opts)
	t.sprite = sprite
}

type CollectedToken struct {
	entity.Token
}

func (t *Token) Collect() entity.CollectedToken {
	return &CollectedToken{
		Token: t,
	}
}

func (t *CollectedToken) Drop(x, y float64) entity.Token {
	t.Token.SetPosition(x, y)
	return t.Token
}
