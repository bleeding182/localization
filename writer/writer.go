package writer

import (
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type Writer interface {
	Tag() string
	Export(locale string, model *LocalizationModel)

	// Convert between %s and %@ and possible other differences between ios/android
	Normalize(string) string

	RegisterCommand(app *kingpin.Application)
}

type LocalizationModel struct {
	Headers *[]string
	Groups  []Group
	Plurals *map[string]QuantityString
}

type Group struct {
	Name    string
	Strings []AndroidString
}

type AndroidString struct {
	Key     string
	Value   string
	Comment string
}

type LocalizedString struct {
	Key            CompositeKey
	Value, Comment string
	Entries        []string
}
