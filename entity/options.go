package entity

type GameOptions struct {
	TechInterview      bool
	DecreaseDifficulty bool
	MusicVolume        float64
	SoundVolume        float64
	ShowFPS            bool
}

func DefaultGameOptions() GameOptions {
	return GameOptions{
		TechInterview:      true,
		DecreaseDifficulty: true,
		MusicVolume:        0.2,
		SoundVolume:        0.2,
	}
}
