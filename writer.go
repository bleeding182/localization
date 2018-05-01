package main

import (
	"fmt"
	"html"
	"os"
	"regexp"
	"sort"
	"sync"
	"time"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

type quantity int

// https://developer.android.com/guide/topics/resources/string-resource.html#Plurals
const (
	zero  quantity = 0 // When the language requires special treatment of the number 0 (as in Arabic).
	one   quantity = 1 // When the language requires special treatment of numbers like one (as with the number 1 in English and most other languages; in Russian, any number ending in 1 but not ending in 11 is in this class).
	two   quantity = 2 // When the language requires special treatment of numbers like two (as with 2 in Welsh, or 102 in Slovenian).
	few   quantity = 3 // When the language requires special treatment of "small" numbers (as with 2, 3, and 4 in Czech; or numbers ending 2, 3, or 4 but not 12, 13, or 14 in Polish).
	many  quantity = 4 // When the language requires special treatment of "large" numbers (as with numbers ending 11-99 in Maltese).
	other quantity = 5 // When the language does not require special treatment of the given quantity (as with all numbers in Chinese, or 42 in English).
)

var quantities = map[quantity]string{
	zero:  "zero",
	one:   "one",
	two:   "two",
	few:   "few",
	many:  "many",
	other: "other",
}

func quantityOf(quantityString string) quantity {
	for q, s := range quantities {
		if s == quantityString {
			return q
		}
	}
	panic(fmt.Sprint("Unknown quantity", quantityString))
}

func (quantity quantity) String() string {
	if quantity < zero || quantity > other {
		panic(fmt.Sprint("Unknown quantity", uint(quantity)))
	}
	return quantities[quantity]
}

type localizedString struct {
	Key            compositeKey
	Value, Comment string
	Entries        []string
}

type quantityString struct {
	Key    string
	Values map[quantity]localizedString
}

type Writer interface {
	Tag() string
	Export(locale string, model *LocalizationModel)

	// Convert between %s and %@ and possible other differences between ios/android
	Normalize(string) string

	registerCommand(app *kingpin.Application)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

var writers = make(map[string]Writer)
var outputFolder *string

var (
	keyColumnName, valueColumnName, commentColumnName *string
)

func registerCommands(app *kingpin.Application) {
	for _, writer := range writers {
		writer.registerCommand(app)
		var tag = writer.Tag()
		command := app.GetCommand(tag)
		if command == nil {
			panic("Command not added under correct tag. Expected " + tag)
		}
	}
	outputFolder = app.Flag("outputFolder", "Set the output directory where the values-* folders will be generated.").Default("exports").String()
	keyColumnName = app.Flag("key", "Override the name of the key column").Default("key").Short('k').String()
	valueColumnName = app.Flag("value", "Override the name of the value column").Default("value").Short('v').String()
	commentColumnName = app.Flag("comment", "Override the name of the comment column").Default("comment").Short('c').String()
}

// I want keys in the form of [group]__?[identifier]__pl_[<one|other|etc>]
// For multi-word groups we can use __ instead to create a long_group__identifier
// 0: original key
// 1: key without (optional) quanity
// 2: group part of key (part until __ or first part until _)
// 3: identifier - part without group or quantity
// 4: optional quantity
var keyRegex = regexp.MustCompile("^((.*?)(?:_{1,2})((?:[a-zA-Z0-9]+_)*[a-zA-Z0-9]*))(?:__pl_(.*))?$")

type compositeKey struct {
	parts []string
}

func (key compositeKey) Original() string {
	return key.parts[0]
}

func (key compositeKey) PlainKey() string {
	return key.parts[1]
}

func (key compositeKey) Group() string {
	return key.parts[2]
}

func (key compositeKey) Identifier() string {
	return key.parts[3]
}

func (key compositeKey) Quantity() string {
	return key.parts[4]
}

type entrySet struct {
	GID    string
	Locale string

	Headers []interface{}
	Values  [][]interface{}
}

type sheet struct {
	GID     string
	Locale  string // e.g. "default" or "en", "en-us"
	Columns map[string]int
	Data    []localizedString
	Plurals map[string]quantityString
	Headers *[]string
	Model   *LocalizationModel
}

type LocalizationModel struct {
	Headers *[]string
	Groups  []Group
	Plurals *map[string]quantityString
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

func (sheet sheet) columnIndex(column string) int {
	index, ok := sheet.Columns[column]
	if !ok {
		return -1
	}
	return index
}

func parse(row []interface{}, index int) string {
	if index >= 0 && index < len(row) {
		return row[index].(string)
	}
	return ""
}

func export(command string, sheetID string, entrySets []*entrySet) (wg *sync.WaitGroup) {
	wg = &sync.WaitGroup{}

	timestamp := time.Now().Format(time.RFC3339)

	sheetChan := make(chan *sheet)
	for _, entrySet := range entrySets {
		go parseEntrySetToSheet(entrySet, sheetChan)
	}

	sheets := make([]*sheet, len(entrySets))
	for i := 0; i < len(entrySets); i++ {
		sheets[i] = <-sheetChan

		sheets[i].Headers = &[]string{
			"Generated by github.com/bleeding182/localization v" + version,
			"Do _not_ modify",
			"https://docs.google.com/spreadsheets/d/" + sheetID + "#gid=" + sheets[i].GID,
			fmt.Sprintf("Last updated at %v", timestamp),
		}

		createSheetModel(command, sheets[i])
	}

	for _, sheet := range sheets {
		for _, w := range writers {
			wg.Add(1)
			go feedWriter(w, sheet, wg)
		}
	}
	return
}

func parseEntrySetToSheet(entrySet *entrySet, sheetChan chan *sheet) {
	fmt.Sprintln("Sheet", entrySet.GID)

	var sheet = &sheet{
		GID:     entrySet.GID,
		Locale:  entrySet.Locale,
		Columns: make(map[string]int),
		Data:    make([]localizedString, 0, len(entrySet.Values)),
		Plurals: make(map[string]quantityString),
	}

	for i, v := range entrySet.Headers {
		header := v.(string)
		sheet.Columns[header] = i
	}

	keyIndex := sheet.columnIndex(*keyColumnName)
	valueIndex := sheet.columnIndex(*valueColumnName)
	commentIndex := sheet.columnIndex(*commentColumnName)

	for _, row := range entrySet.Values {
		key := parse(row, keyIndex)
		if key == "" {
			continue
		}

		compositeKey := compositeKey{keyRegex.FindStringSubmatch(key)}

		s := localizedString{
			Key:     compositeKey,
			Value:   parse(row, valueIndex),
			Comment: parse(row, commentIndex),
			Entries: make([]string, len(row)),
		}
		for i, c := range row {
			s.Entries[i] = c.(string)
		}
		sheet.Data = append(sheet.Data, s)

		if compositeKey.Quantity() != "" {
			key := compositeKey.PlainKey()

			plural, ok := sheet.Plurals[key]
			if !ok {
				plural = quantityString{
					Key:    key,
					Values: make(map[quantity]localizedString),
				}
				sheet.Plurals[key] = plural
			}

			plural.Values[quantityOf(compositeKey.Quantity())] = s
		}
	}

	sort.Slice(sheet.Data, func(i, j int) bool {
		return sheet.Data[i].Key.Original() < sheet.Data[j].Key.Original()
	})

	sheetChan <- sheet
}

func createSheetModel(tag string, sheet *sheet) {

	overrideIndex, ok := sheet.Columns[tag]
	if !ok {
		overrideIndex = -1
	}

	var writer = writers[tag]

	model := &LocalizationModel{
		Headers: sheet.Headers,
		Groups:  make([]Group, 0),
		Plurals: &sheet.Plurals,
	}

	var group Group
	for _, ls := range sheet.Data {
		if group.Name != ls.Key.Group() {
			if group.Name != "" {
				model.Groups = append(model.Groups, group)
			}
			group = Group{
				Name:    ls.Key.Group(),
				Strings: make([]AndroidString, 0),
			}
		}
		var value string
		if overrideIndex >= 0 && overrideIndex < len(ls.Entries) && ls.Entries[overrideIndex] != "" {
			value = ls.Entries[overrideIndex]
		} else {
			value = html.EscapeString(writer.Normalize(ls.Value))
		}

		group.Strings = append(group.Strings, AndroidString{ls.Key.Original(), value, ls.Comment})
	}
	if group.Name != "" {
		model.Groups = append(model.Groups, group)
	}

	sheet.Model = model
}

func feedWriter(w Writer, sheet *sheet, wg *sync.WaitGroup) {
	defer wg.Done()
	w.Export(sheet.Locale, sheet.Model)
}

func openFile(folder string, name string) *os.File {
	foldername := fmt.Sprintf("%v", folder)
	os.MkdirAll(foldername, os.ModePerm)

	filename := fmt.Sprintf("%v/%v", foldername, name)
	f, err := os.Create(filename)
	check(err)
	return f
}
