package scenes

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"time"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/fogleman/primitive/primitive"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font"
)

type gameOverScene struct {
	ctx        context.Context
	game       entity.Game
	ui         *ebitenui.UI
	fontFace10 font.Face
	questions  entity.QuestionsForInterview
	next       entity.Scene

	whiteSub   *ebiten.Image
	model      *primitive.Model
	tris       []Tri
	stepsLeft  int
	bg         color.NRGBA
	imgW, imgH int
}

type Tri struct {
	x0, x1, x2 float32
	y0, y1, y2 float32
	r, g, b, a float32
}

func NewGameOverScene(ctx context.Context, game entity.Game, fontFace10 font.Face, questions entity.QuestionsForInterview, stats entity.Stats) entity.Scene {
	s := &gameOverScene{
		ctx:        ctx,
		game:       game,
		fontFace10: fontFace10,
		questions:  questions,
	}

	fontSrc, err := text.NewGoTextFaceSource(bytes.NewReader(assets.StampatelloFacetoKernTTF))
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	face := text.Face(&text.GoTextFace{Source: fontSrc, Size: 18})

	inGameDays := int(stats.DurationInGame.Hours() / 24)
	realMinutes := int(stats.RealTimePlayed.Minutes())

	type statRow struct{ label, value string }
	rows := []statRow{
		{fmt.Sprintf("%d in-game days in %d minutes of real time.", inGameDays, realMinutes), ""},
		{"", ""},
		{"Current balance:", moneyToString(stats.CurrentBalance)},
		{"Monthly salary:", moneyToString(stats.MonthlySalary)},
		{"Earned:", moneyToString(stats.Earned)},
		{"Spent:", moneyToString(stats.Spent)},
		{"", ""},
		{"Questions answered:", fmt.Sprintf("%d", stats.QuestionsAnswered)},
		{"Offers accepted:", fmt.Sprintf("%d", stats.OffersAccepted)},
		{"Offers rejected:", fmt.Sprintf("%d", stats.OffersRejected)},
		{"Offers received:", fmt.Sprintf("%d", stats.OffersReceived)},
	}

	rootContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
	)

	centerContainer := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Spacing(8),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(40)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
				HorizontalPosition: widget.AnchorLayoutPositionCenter,
				VerticalPosition:   widget.AnchorLayoutPositionCenter,
			}),
		),
	)

	addText := func(c *widget.Container, t string, pos widget.GridLayoutPosition) {
		c.AddChild(widget.NewText(
			widget.TextOpts.Text(t, &face, color.White),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.GridLayoutData{
					HorizontalPosition: pos,
					VerticalPosition:   widget.GridLayoutPositionCenter,
				}),
			),
		))
	}

	for _, row := range rows {
		if row.label == "" && row.value == "" {
			centerContainer.AddChild(widget.NewText(
				widget.TextOpts.Text(" ", &face, color.White),
				widget.TextOpts.WidgetOpts(
					widget.WidgetOpts.LayoutData(widget.RowLayoutData{}),
				),
			))
			continue
		}
		rowContainer := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewGridLayout(
				widget.GridLayoutOpts.Columns(2),
				widget.GridLayoutOpts.Stretch([]bool{true, false}, nil),
				widget.GridLayoutOpts.Spacing(24, 0),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
				widget.WidgetOpts.MinSize(400, 0),
			),
		)
		addText(rowContainer, row.label, widget.GridLayoutPositionStart)
		addText(rowContainer, row.value, widget.GridLayoutPositionEnd)
		centerContainer.AddChild(rowContainer)
	}

	buttonImage := &widget.ButtonImage{
		Idle:    euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 60, G: 60, B: 80, A: 255}, color.NRGBA{R: 120, G: 120, B: 160, A: 255}, 2),
		Hover:   euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 80, G: 80, B: 110, A: 255}, color.NRGBA{R: 160, G: 160, B: 200, A: 255}, 2),
		Pressed: euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 40, G: 40, B: 60, A: 255}, color.NRGBA{R: 100, G: 100, B: 140, A: 255}, 2),
	}

	btn := widget.NewButton(
		widget.ButtonOpts.Image(buttonImage),
		widget.ButtonOpts.Text("Main Menu", &face, &widget.ButtonTextColor{
			Idle:  color.White,
			Hover: color.RGBA{200, 220, 255, 255},
		}),
		widget.ButtonOpts.TextPadding(&widget.Insets{Left: 30, Right: 30, Top: 8, Bottom: 8}),
		widget.ButtonOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
		widget.ButtonOpts.ClickedHandler(func(args *widget.ButtonClickedEventArgs) {
			s.next = NewMainMenuScene(ctx, s.game, s.questions)
		}),
	)

	centerContainer.AddChild(widget.NewText(
		widget.TextOpts.Text(" ", &face, color.White),
	))
	centerContainer.AddChild(btn)

	rootContainer.AddChild(centerContainer)

	s.ui = &ebitenui.UI{Container: rootContainer}

	const steps = 3000

	whiteImage := ebiten.NewImage(3, 3)
	whiteImage.Fill(color.White)
	s.whiteSub = whiteImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)

	imgSrc, _, err := image.Decode(bytes.NewReader(assets.GameOverPNG))
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	src := ebiten.NewImageFromImage(imgSrc)

	b := src.Bounds()
	bg := primitive.MakeColor(primitive.AverageImageColor(src))
	model := primitive.NewModel(src, bg, 256, 4)

	s.model = model
	s.bg = bg.NRGBA()
	s.stepsLeft = steps
	s.imgW = b.Max.X
	s.imgH = b.Max.Y

	return s
}

func (s *gameOverScene) Update() (entity.Scene, error) {
	s.ui.Update()
	if s.next != nil {
		return s.next, nil
	}

	deadline := time.Now().Add(8 * time.Millisecond)
	for s.stepsLeft > 0 && time.Now().Before(deadline) {
		s.model.Step(primitive.ShapeTypeTriangle, 128, 0)
		s.stepsLeft--
		shape := s.model.Shapes[len(s.model.Shapes)-1]
		if t, ok := shape.(*primitive.Triangle); ok {
			c := s.model.Colors[len(s.model.Colors)-1]
			s.tris = append(s.tris, Tri{
				x0: float32(t.X1), y0: float32(t.Y1),
				x1: float32(t.X2), y1: float32(t.Y2),
				x2: float32(t.X3), y2: float32(t.Y3),
				r: float32(c.R) / 255,
				g: float32(c.G) / 255,
				b: float32(c.B) / 255,
				a: float32(c.A) / 255,
			})
		}
	}

	return nil, nil
}

func (s *gameOverScene) Draw(screen *ebiten.Image) {
	winW, winH := s.game.WindowSize()
	scaleX := float32(winW) / float32(s.imgW)
	scaleY := float32(winH) / float32(s.imgH)

	for _, t := range s.tris {
		var p vector.Path
		p.MoveTo(t.x0*scaleX, t.y0*scaleY)
		p.LineTo(t.x1*scaleX, t.y1*scaleY)
		p.LineTo(t.x2*scaleX, t.y2*scaleY)
		p.Close()

		vs, is := p.AppendVerticesAndIndicesForFilling(nil, nil)
		for i := range vs {
			vs[i].SrcX = 1
			vs[i].SrcY = 1
			vs[i].ColorR = t.r
			vs[i].ColorG = t.g
			vs[i].ColorB = t.b
			vs[i].ColorA = t.a
		}
		screen.DrawTriangles(vs, is, s.whiteSub, nil)
	}
	s.ui.Draw(screen)
}

func (s *gameOverScene) Name() string { return "Game Over" }
