package character

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"time"

	"github.com/dqso/ludum-dare-59/assets"
	"github.com/dqso/ludum-dare-59/entity"
	"github.com/hajimehoshi/ebiten/v2"
)

const targetH float64 = 45.0

type Character struct {
	x, y            float64
	w, h            float64
	pivotX, pivotY  float64
	speed           float64
	sprite          *ebiten.Image
	role            entity.CharacterRole
	focused         bool
	questions       []entity.Question
	playerPoints    int
	deadline        time.Time
	company         string
	prev            entity.Character
	next            entity.Character
	interviewResult entity.InterviewResult
}

func NewCharacter(role entity.CharacterRole, x, y float64, questions []entity.Question, lifetime time.Duration, company string) (*Character, error) {
	var asset []byte
	switch role {
	case entity.CharacterRoleRecruiter:
		asset = assets.RecruiterPNG
	case entity.CharacterRoleEngineer:
		asset = assets.EngineerPNG
	case entity.CharacterRoleFounder:
		asset = assets.OwnerPNG
	default:
		return nil, fmt.Errorf("unknown character role: %v", role)
	}
	img, _, err := image.Decode(bytes.NewReader(asset))
	if err != nil {
		return nil, fmt.Errorf("failed to decode sprite: %v", err)
	}
	src := ebiten.NewImageFromImage(img)

	scale := targetH / float64(src.Bounds().Dy())
	w := float64(src.Bounds().Dx()) * scale

	sprite := ebiten.NewImage(int(w), int(targetH))
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(scale, scale)
	sprite.DrawImage(src, op)

	if lifetime == 0 {
		lifetime = time.Hour * 24 * 256 * 5
	}
	c := &Character{
		x:         0,
		y:         0,
		w:         w,
		h:         targetH,
		pivotX:    0,
		pivotY:    0,
		speed:     1.0,
		sprite:    sprite,
		role:      role,
		questions: questions,
		deadline:  time.Now().Add(lifetime),
		company:   company,
	}
	c.Move(x, y)

	return c, nil
}

func (c Character) Role() entity.CharacterRole {
	return c.role
}

func (c Character) X() float64                                        { return c.x }
func (c Character) Y() float64                                        { return c.y }
func (c Character) TopLeftX() float64                                 { return c.x - c.w/2 }
func (c Character) TopLeftY() float64                                 { return c.y - c.h/2 }
func (c Character) Speed() float64                                    { return c.speed }
func (c Character) PivotX() float64                                   { return c.pivotX }
func (c Character) PivotY() float64                                   { return c.pivotY }
func (c Character) PivotRadius() float64                              { return c.w / 2 }
func (c Character) Width() float64                                    { return c.w }
func (c Character) Height() float64                                   { return c.h }
func (c *Character) SetFocus(focus bool)                              { c.focused = focus }
func (c Character) IsFocused() bool                                   { return c.focused }
func (c Character) Deadline() time.Time                               { return c.deadline }
func (c *Character) UpdateDeadline(delay time.Duration)               { c.deadline = time.Now().Add(delay) }
func (c Character) Company() string                                   { return c.company }
func (c *Character) SetCompany(company string)                        { c.company = company }
func (c Character) InterviewResult() entity.InterviewResult           { return c.interviewResult }
func (c *Character) SetInterviewResult(result entity.InterviewResult) { c.interviewResult = result }
func (c Character) PlayerPoints() int                                 { return c.playerPoints }
func (c *Character) SetPlayerPoints(playerPoints int)                 { c.playerPoints = playerPoints }

func (c *Character) Move(dx, dy float64) {
	c.x += dx
	c.y += dy
	c.pivotX += dx
	c.pivotY += dy
}

func (c Character) GetQuestion() entity.Question {
	if len(c.questions) > 0 {
		return c.questions[0]
	}
	return nil
}

func (c *Character) AnswerTheQuestion(questionsDatabase entity.QuestionsDatabase, answer entity.Answer) {
	if len(c.questions) == 0 {
		return
	}

	var question entity.Question
	question, c.questions = c.questions[0], c.questions[1:]
	points, err := questionsDatabase.Match(question, answer)
	if err != nil {
		log.Printf("failed to answer the question %q using the answer %q: %v", question, answer, err)
		return
	}

	c.playerPoints += points

	logAboutPoints := "earned nothing"
	if points > 0 {
		logAboutPoints = fmt.Sprintf("earned %d points", points)
	} else if points < 0 {
		logAboutPoints = fmt.Sprintf("lost %d points", max(points, -points))
	}
	log.Printf("Player has answered the question %q using the answer %q and %s.",
		question.Question(), answer.Answer(), logAboutPoints)
}

func (c Character) Draw(screen *ebiten.Image, op *ebiten.DrawImageOptions) {
	screen.DrawImage(c.sprite, op)
}
