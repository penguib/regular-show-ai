package models

import "strings"

type Character int

const (
	Mordecai Character = 1 << iota
	Rigby
	Pops
	Benson
	MuscleMan
	Doe
)

type Speech struct {
	Character Character
	Content   string
	AudioFile string `json:"audio_file"`
}

type Scene struct {
	ID           int
	Dirty        bool `json:"dirty,omitempty"`
	Characters   int
	Conversation []Speech
}

func GetCharacterByName(name string) Character {
	name = strings.ToLower(name)
	switch name {
	case "mordecai":
		return Mordecai
	case "rigby":
		return Rigby
	case "pops":
		return Pops
	case "benson":
		return Benson
	case "muscle man":
		return MuscleMan
	}

	return Doe
}
