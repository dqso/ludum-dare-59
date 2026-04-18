package assets

import (
	_ "embed"
)

var (
	//go:embed answers.csv
	AnswersCSV []byte

	//go:embed mainchar.png
	MainCharPNG []byte

	//go:embed girl.png
	GirlPNG []byte

	//go:embed overdozesans.ttf
	OverdoseSansTTF []byte

	//go:embed StampatelloFacetoKern.ttf
	StampatelloFacetoKernTTF []byte
)
