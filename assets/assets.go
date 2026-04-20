package assets

import (
	_ "embed"
)

var (
	//go:embed GameOver.png
	GameOverPNG []byte

	//go:embed recruiter.csv
	QuestionsRecruiterCSV []byte

	//go:embed engineer.csv
	QuestionsEngineerCSV []byte

	//go:embed mainchar.png
	MainCharPNG []byte

	//go:embed recruiter.png
	RecruiterPNG []byte

	//go:embed engineer.png
	EngineerPNG []byte

	//go:embed owner.png
	OwnerPNG []byte

	//go:embed overdozesans.ttf
	OverdoseSansTTF []byte

	//go:embed StampatelloFacetoKern.ttf
	StampatelloFacetoKernTTF []byte
)
