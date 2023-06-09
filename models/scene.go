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

func GetCharacterVoiceUUID(character Character) string {
	switch character {
	case Mordecai:
		return "7e67372e-d875-419b-8e91-15aeefdac0a9"
	case Rigby:
		return "d871d692-7c87-42e4-810b-5acb3083cefd"
	case Pops:
		return "54cf7ae5-f71d-4b31-accc-32d2594e4250"
	case Benson:
		return "af4a692e-9c7c-42e8-b3f7-db6e0f68bfd6"
	case MuscleMan:
		return "33f4232f-00dd-4239-863d-d71454ba2475"
	}
	return ""
}
