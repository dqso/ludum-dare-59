package assets

import (
	_ "embed"
)

var (
	//go:embed female-first-names.txt
	FemaleFirstNamesTxt []byte

	//go:embed male-first-names.txt
	MaleFirstNamesTxt []byte

	//go:embed company-names.txt
	CompanyNamesTxt []byte

	//go:embed GameOver.png
	GameOverPNG []byte

	//go:embed win.png
	WinPNG []byte

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

	//go:embed StampatelloFacetoKern.ttf
	StampatelloFacetoKernTTF []byte

	//go:embed CascadiaMono.ttf
	CascadiaMonoTTF []byte
)
