package scenes

import (
	"bytes"
	"context"
	"fmt"
	"image/color"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/ebitenui/ebitenui"
	euiimage "github.com/ebitenui/ebitenui/image"
	"github.com/ebitenui/ebitenui/widget"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type mainMenuScreen int

const (
	mainMenuScreenMain mainMenuScreen = iota
	mainMenuScreenOptions
	mainMenuScreenHelp
)

type mainMenuScene struct {
	ctx        context.Context
	game       entity.Game
	questions  entity.QuestionsForInterview
	firstNames entity.FirstNamesDatabase

	options entity.GameOptions

	screen                mainMenuScreen
	mainUI                *ebitenui.UI
	optionsUI             *ebitenui.UI
	helpUI                *ebitenui.UI
	techInterviewBtn      *widget.Button
	decreaseDifficultyBtn *widget.Button
	musicVolumeBtn        *widget.Button
	showFPSBtn            *widget.Button

	next entity.Scene
	quit bool
}

func NewMainMenuScene(ctx context.Context, game entity.Game, questions entity.QuestionsForInterview, firstNames entity.FirstNamesDatabase) entity.Scene {
	s := &mainMenuScene{
		ctx:        ctx,
		game:       game,
		questions:  questions,
		firstNames: firstNames,
		options: entity.GameOptions{
			TechInterview:      true,
			DecreaseDifficulty: true,
			MusicVolume:        0.2,
		},
	}

	fontSrc, err := text.NewGoTextFaceSource(bytes.NewReader(assets.StampatelloFacetoKernTTF))
	if err != nil {
		return NewErrorScene(ctx, game, err)
	}
	face := text.Face(&text.GoTextFace{Source: fontSrc, Size: 18})
	faceSmall := text.Face(&text.GoTextFace{Source: fontSrc, Size: 14})

	buttonImage := &widget.ButtonImage{
		Idle:    euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 60, G: 60, B: 80, A: 255}, color.NRGBA{R: 120, G: 120, B: 160, A: 255}, 2),
		Hover:   euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 80, G: 80, B: 110, A: 255}, color.NRGBA{R: 160, G: 160, B: 200, A: 255}, 2),
		Pressed: euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 40, G: 40, B: 60, A: 255}, color.NRGBA{R: 100, G: 100, B: 140, A: 255}, 2),
	}

	btnTextColor := &widget.ButtonTextColor{
		Idle:  color.White,
		Hover: color.RGBA{200, 220, 255, 255},
	}

	newButton := func(label string, handler func()) *widget.Button {
		return widget.NewButton(
			widget.ButtonOpts.Image(buttonImage),
			widget.ButtonOpts.Text(label, &face, btnTextColor),
			widget.ButtonOpts.TextPadding(&widget.Insets{Left: 40, Right: 40, Top: 10, Bottom: 10}),
			widget.ButtonOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
				widget.WidgetOpts.MinSize(220, 0),
			),
			widget.ButtonOpts.ClickedHandler(func(_ *widget.ButtonClickedEventArgs) {
				handler()
			}),
		)
	}

	makeCenterContainer := func() *widget.Container {
		return widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewRowLayout(
				widget.RowLayoutOpts.Direction(widget.DirectionVertical),
				widget.RowLayoutOpts.Spacing(16),
				widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(60)),
			)),
			widget.ContainerOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.AnchorLayoutData{
					HorizontalPosition: widget.AnchorLayoutPositionCenter,
					VerticalPosition:   widget.AnchorLayoutPositionCenter,
				}),
			),
		)
	}

	makeRoot := func(inner *widget.Container) *widget.Container {
		root := widget.NewContainer(
			widget.ContainerOpts.Layout(widget.NewAnchorLayout()),
		)
		root.AddChild(inner)
		return root
	}

	addTitle := func(c *widget.Container, title string) {
		c.AddChild(widget.NewText(
			widget.TextOpts.Text(title, &face, color.RGBA{180, 200, 255, 255}),
			widget.TextOpts.WidgetOpts(
				widget.WidgetOpts.LayoutData(widget.RowLayoutData{
					Position: widget.RowLayoutPositionCenter,
				}),
			),
		))
		c.AddChild(widget.NewText(
			widget.TextOpts.Text(" ", &face, color.White),
		))
	}

	// --- Main screen ---
	mainC := makeCenterContainer()
	addTitle(mainC, "Token Collector")

	mainC.AddChild(newButton("New Game", func() {
		s.next = NewGenerateBattlefieldScene(s.ctx, s.game, s.questions, s.firstNames, s.options)
	}))
	mainC.AddChild(newButton("Options", func() {
		s.screen = mainMenuScreenOptions
	}))
	mainC.AddChild(newButton("Help", func() {
		s.screen = mainMenuScreenHelp
	}))
	mainC.AddChild(newButton("Quit", func() {
		s.quit = true
	}))

	s.mainUI = &ebitenui.UI{Container: makeRoot(mainC)}

	// --- Options screen ---
	optC := makeCenterContainer()
	addTitle(optC, "Options")

	s.techInterviewBtn = newButton(s.techInterviewLabel(), func() {
		s.options.TechInterview = !s.options.TechInterview
	})
	optC.AddChild(s.techInterviewBtn)

	s.decreaseDifficultyBtn = newButton(s.decreaseDifficultyLabel(), func() {
		s.options.DecreaseDifficulty = !s.options.DecreaseDifficulty
	})
	optC.AddChild(s.decreaseDifficultyBtn)

	optC.AddChild(widget.NewText(
		widget.TextOpts.Text("Removes other difficulties, e.g. no need to pay rent", &faceSmall, color.NRGBA{R: 180, G: 180, B: 180, A: 200}),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	))

	s.musicVolumeBtn = newButton(s.musicVolumeLabel(), func() {
		levels := musicVolumeLevels()
		for i, v := range levels {
			if v == s.options.MusicVolume {
				s.options.MusicVolume = levels[(i+1)%len(levels)]
				return
			}
		}
		s.options.MusicVolume = levels[1]
	})
	optC.AddChild(s.musicVolumeBtn)

	s.showFPSBtn = newButton(s.showFPSLabel(), func() {
		s.options.ShowFPS = !s.options.ShowFPS
	})
	optC.AddChild(s.showFPSBtn)

	optC.AddChild(newButton("Back", func() {
		s.screen = mainMenuScreenMain
	}))

	s.optionsUI = &ebitenui.UI{Container: makeRoot(optC)}

	// --- Help screen ---
	helpC := makeCenterContainer()
	addTitle(helpC, "How to Play")

	helpText := "Move: WASD / Arrow keys.\n" +
		"Collect tokens: approach and press E / drop 1–9.\n" +
		"Interview: approach a character and answer questions with 1–9.\n\n" +
		"Earn money by passing interviews and accepting job offers.\n\n" +
		"Your financial cushion decreases over time - keep grinding to stay afloat!\n\n" +
		"The interview consists of three stages: recruiter, tech specialist, owner.\n\n" +
		"The game ends when you run out of money. Good luck!\n\n" +
		"Ludum Dare 59. Theme is Signal.\n\n" +
		"Author: github.com/dqso"

	scrollC := widget.NewContainer(
		widget.ContainerOpts.Layout(widget.NewRowLayout(
			widget.RowLayoutOpts.Direction(widget.DirectionVertical),
			widget.RowLayoutOpts.Padding(widget.NewInsetsSimple(10)),
		)),
		widget.ContainerOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{Stretch: true}),
		),
	)
	scrollC.AddChild(widget.NewText(
		widget.TextOpts.Text(helpText, &faceSmall, color.White),
		widget.TextOpts.WidgetOpts(
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{}),
		),
	))

	trackImage := &widget.ScrollContainerImage{
		Idle:     euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 30, G: 30, B: 50, A: 220}, color.NRGBA{R: 80, G: 80, B: 120, A: 255}, 1),
		Disabled: euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 30, G: 30, B: 50, A: 220}, color.NRGBA{R: 80, G: 80, B: 120, A: 255}, 1),
		Mask:     euiimage.NewBorderedNineSliceColor(color.NRGBA{R: 30, G: 30, B: 50, A: 220}, color.NRGBA{R: 80, G: 80, B: 120, A: 255}, 1),
	}
	scrollContainer := widget.NewScrollContainer(
		widget.ScrollContainerOpts.Content(scrollC),
		widget.ScrollContainerOpts.Image(trackImage),
		widget.ScrollContainerOpts.StretchContentWidth(),
		widget.ScrollContainerOpts.WidgetOpts(
			widget.WidgetOpts.MinSize(400, 280),
			widget.WidgetOpts.LayoutData(widget.RowLayoutData{
				Position: widget.RowLayoutPositionCenter,
			}),
		),
	)
	helpC.AddChild(scrollContainer)
	helpC.AddChild(newButton("Back", func() {
		s.screen = mainMenuScreenMain
	}))

	s.helpUI = &ebitenui.UI{Container: makeRoot(helpC)}

	return s
}

func (s *mainMenuScene) techInterviewLabel() string {
	state := "off"
	if s.options.TechInterview {
		state = "on"
	}
	return fmt.Sprintf("Tech Interview [%s]", state)
}

func (s *mainMenuScene) decreaseDifficultyLabel() string {
	state := "off"
	if s.options.DecreaseDifficulty {
		state = "on"
	}
	return fmt.Sprintf("Decrease Difficulty [%s]", state)
}

func (s *mainMenuScene) showFPSLabel() string {
	state := "off"
	if s.options.ShowFPS {
		state = "on"
	}
	return fmt.Sprintf("Show FPS [%s]", state)
}

func musicVolumeLevels() []float64 { return []float64{0, 0.1, 0.2, 0.5, 0.7, 1.0} }

func (s *mainMenuScene) musicVolumeLabel() string {
	labels := map[float64]string{0: "Off", 0.1: "Very Low", 0.2: "Low", 0.5: "Medium", 0.7: "High", 1.0: "Max"}
	name, ok := labels[s.options.MusicVolume]
	if !ok {
		name = "Custom"
	}
	return fmt.Sprintf("Music Volume [%s]", name)
}

func (s *mainMenuScene) Update() (entity.Scene, error) {
	if s.quit {
		return nil, ebiten.Termination
	}
	if s.next != nil {
		return s.next, nil
	}

	escPressed := inpututil.IsKeyJustPressed(ebiten.KeyEscape)

	switch s.screen {
	case mainMenuScreenMain:
		if escPressed {
			s.quit = true
		}
		s.mainUI.Update()

	case mainMenuScreenOptions:
		if escPressed {
			s.screen = mainMenuScreenMain
		}
		s.techInterviewBtn.Text().Label = s.techInterviewLabel()
		s.decreaseDifficultyBtn.Text().Label = s.decreaseDifficultyLabel()
		s.musicVolumeBtn.Text().Label = s.musicVolumeLabel()
		s.showFPSBtn.Text().Label = s.showFPSLabel()
		s.optionsUI.Update()

	case mainMenuScreenHelp:
		if escPressed {
			s.screen = mainMenuScreenMain
		}
		s.helpUI.Update()
	}

	return nil, nil
}

func (s *mainMenuScene) Draw(screen *ebiten.Image) {
	screen.Fill(color.NRGBA{R: 15, G: 15, B: 30, A: 255})

	switch s.screen {
	case mainMenuScreenMain:
		s.mainUI.Draw(screen)
	case mainMenuScreenOptions:
		s.optionsUI.Draw(screen)
	case mainMenuScreenHelp:
		s.helpUI.Draw(screen)
	}
}

func (s *mainMenuScene) Name() string { return "Main Menu" }
