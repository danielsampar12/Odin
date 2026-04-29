package companions

import "strings"

type Companion struct {
	Key         string
	Name        string
	Description string
}

var Registry = []Companion{
	{Key: "baldur", Name: "Baldur", Description: "Pragmatic pair programmer"},
	{Key: "tyr", Name: "Tyr", Description: "Architect and reviewer"},
	{Key: "thor", Name: "Thor", Description: "Fast implementer"},
	{Key: "loki", Name: "Loki", Description: "Creative brainstormer"},
	{Key: "freya", Name: "Freya", Description: "Beginner-friendly teacher"},
	{Key: "hephaestus", Name: "Hephaestus", Description: "Infra and local setup expert"},
}

func DefaultForProfile(profile string) Companion {
	if strings.EqualFold(profile, "beginner") {
		return Companion{Key: "freya", Name: "Freya", Description: "Beginner-friendly teacher"}
	}

	return Companion{Key: "baldur", Name: "Baldur", Description: "Pragmatic pair programmer"}
}
