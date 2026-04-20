package entity

import "github.com/dqso/ludum-dare-59/assets"

func GetMusicPlaylist() [][]byte {
	return [][]byte{assets.Music1MP3, assets.Music2MP3, assets.Music3MP3}
}
