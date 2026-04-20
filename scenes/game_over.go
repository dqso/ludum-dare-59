package scenes

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	_ "image/png"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font"
)

type gameOverScene struct {
	ctx             context.Context
	game            entity.Game
	bg              *ebiten.Image
	bgBlurred       *ebiten.Image
	ui              *ebitenui.UI
	_type           TypeGameOver
	fontFace10      font.Face
	questions       entity.QuestionsForInterview
	firstNames      entity.FirstNamesDatabase
	centerContainer *widget.Container
	next            entity.Scene
}

type TypeGameOver string

const (
	typeGameOverWin  TypeGameOver = "win"
	typeGameOverLose TypeGameOver = "loss"
)

func NewGameOverScene(ctx context.Context, _type TypeGameOver, game entity.Game, fontFace10 font.Face, questions entity.QuestionsForInterview, firstNames entity.FirstNamesDatabase, stats entity.Stats) entity.Scene {
	title, bgBytes := "You lose.", assets.GameOverPNG
	if _type == typeGameOverWin {
		title, bgBytes = "You win.", assets.WinPNG
	}
	bgImg, _, err := image.Decode(bytes.NewReader(bgBytes))
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}

	bgFull := ebiten.NewImageFromImage(bgImg)
	const blurScale = 8
	bw, bh := bgFull.Bounds().Dx()/blurScale, bgFull.Bounds().Dy()/blurScale
	bgBlurred := ebiten.NewImage(bw, bh)
	blurOp := &ebiten.DrawImageOptions{}
	blurOp.GeoM.Scale(1.0/blurScale, 1.0/blurScale)
	bgBlurred.DrawImage(bgFull, blurOp)

	s := &gameOverScene{
		ctx:        ctx,
		game:       game,
		_type:      _type,
		firstNames: firstNames,
		bg:         bgFull,
		bgBlurred:  bgBlurred,
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
		{title, ""},
		{"", ""},
		{fmt.Sprintf("%d in-game days in %d minutes of real time.", inGameDays, realMinutes), ""},
		{"", ""},
		{"Current balance:", moneyToString(stats.CurrentBalance)},
		{"Monthly salary:", moneyToString(stats.MonthlySalary)},
		{"Earned:", moneyToString(stats.Earned)},
		{"Spent:", moneyToString(stats.Spent)},
		{"", ""},
		{"Questions answered:", fmt.Sprintf("%d", stats.QuestionsAnswered)},
		{"Offers received:", fmt.Sprintf("%d", stats.OffersReceived)},
		{"Offers accepted:", fmt.Sprintf("%d", stats.OffersAccepted)},
		{"Offers rejected:", fmt.Sprintf("%d", stats.OffersRejected)},
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
			s.next = NewMainMenuScene(ctx, s.game, s.questions, s.firstNames)
		}),
	)

	centerContainer.AddChild(widget.NewText(
		widget.TextOpts.Text(" ", &face, color.White),
	))
	centerContainer.AddChild(btn)

	s.centerContainer = centerContainer
	rootContainer.AddChild(centerContainer)

	s.ui = &ebitenui.UI{Container: rootContainer}

	return s
}

func (s *gameOverScene) Update() (entity.Scene, error) {
	s.ui.Update()
	if s.next != nil {
		return s.next, nil
	}

	return nil, nil
}

func (s *gameOverScene) Draw(screen *ebiten.Image) {
	winW, winH := s.game.WindowSize()
	scaleX := float64(winW) / float64(s.bg.Bounds().Dx())
	scaleY := float64(winH) / float64(s.bg.Bounds().Dy())
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scaleX, scaleY)
	screen.DrawImage(s.bg, op)

	// Frosted glass panel behind the stats container.
	r := s.centerContainer.GetWidget().Rect
	if !r.Empty() && s._type != typeGameOverLose {
		panelX, panelY := float64(r.Min.X), float64(r.Min.Y)
		panelW, panelH := float64(r.Dx()), float64(r.Dy())
		// Blurred bg: scale bgBlurred (1/8 of original) to match the panel area.
		bw := float64(s.bgBlurred.Bounds().Dx())
		bh := float64(s.bgBlurred.Bounds().Dy())
		srcX := panelX / float64(winW) * bw
		srcY := panelY / float64(winH) * bh
		srcW := panelW / float64(winW) * bw
		srcH := panelH / float64(winH) * bh
		blurOp := &ebiten.DrawImageOptions{}
		blurOp.GeoM.Scale(panelW/srcW, panelH/srcH)
		blurOp.GeoM.Translate(panelX, panelY)
		screen.DrawImage(s.bgBlurred.SubImage(image.Rect(
			int(srcX), int(srcY), int(srcX+srcW), int(srcY+srcH),
		)).(*ebiten.Image), blurOp)
		// Dark tint over the blurred region.
		overlay := ebiten.NewImage(int(panelW), int(panelH))
		overlay.Fill(color.NRGBA{0, 0, 0, 140})
		overlayOp := &ebiten.DrawImageOptions{}
		overlayOp.GeoM.Translate(panelX, panelY)
		screen.DrawImage(overlay, overlayOp)
	}

	s.ui.Draw(screen)
}

func (s *gameOverScene) Name() string { return "Game Over" }
