package main

import (
	"fmt"
	"html"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/bleeding182/localization/writer"
	"github.com/bleeding182/localization/writer/android"
	"github.com/bleeding182/localization/writer/ios"
	"gopkg.in/alecthomas/kingpin.v2"
)

// go build -ldflags "-X main.version={version}"
var version string

const appDescription = `
A CLI-Tool to export localized strings from your Google Sheets.

Your sheet should contain the following columns: [key, value, android, ios, comment] These default names can be overriden.

Keys:
    The keys used for your strings must match "<group>_<identifier>". If yor group name consists of multiple parts you can use "__" to mark it accordingly. Groups are used to group strings together.

Plurals:
	Plurals must be marked by the "__pl_<zero|one|two|few|many|other>" suffix on your key. If supported by the export target, they will be exported and grouped accordingly.
	
Values:
    Values may be escaped by the target platform or modified in some other way. If you want to override a value you can place it in a column of the target platforms name.

For more information please also check https://github.com/bleeding182/localization where this app is published under the MIT Open Source License.
`

var (
	app = kingpin.New("localization", appDescription).Version(version)
	// verbose = app.Flag("verbose", "Verbose logs. Use this to debug potential errors.").Bool()
	sheetID = app.Flag("sheetID", "ID of the spreadsheet to use.").Short('s').Required().String()
)

func main() {
	RegisterCommands(app)
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	resp, err := loadSpreadSheet()

	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}

	entrySets := make([]*EntrySet, len(resp))
	for i, data := range resp {
		sheet := data.sheet
		entrySets[i] = &EntrySet{
			GID:     data.sheetID,
			Locale:  strings.Split(sheet.Range, "!")[0],
			Headers: sheet.Values[:1][0],
			Values:  sheet.Values[1:],
		}
	}

	wg := Export(command, *sheetID, entrySets)

	wg.Wait()
}

var outputFolder *string

var (
	keyColumnName, valueColumnName, commentColumnName *string
)

var Writers = map[string]writer.Writer{
	"ios":     ios.IOSWriter{},
	"android": android.AndroidWriter{},
}

type EntrySet struct {
	GID    string
	Locale string

	Headers []interface{}
	Values  [][]interface{}
}

func RegisterCommands(app *kingpin.Application) {
	for _, writer := range Writers {
		writer.RegisterCommand(app)
		var tag = writer.Tag()
		command := app.GetCommand(tag)
		if command == nil {
			panic("Command not added under correct tag. Expected " + tag)
		}
	}
	keyColumnName = app.Flag("key", "Override the name of the key column").Default("key").Short('k').String()
	valueColumnName = app.Flag("value", "Override the name of the value column").Default("value").Short('v').String()
	commentColumnName = app.Flag("comment", "Override the name of the comment column").Default("comment").Short('c').String()
}

type sheet struct {
	GID     string
	Locale  string // e.g. "default" or "en", "en-us"
	Columns map[string]int
	Data    []writer.LocalizedString
	Plurals map[string]writer.QuantityString
	Headers *[]string
	Model   *writer.LocalizationModel
}

func (sheet sheet) columnIndex(column string) int {
	index, ok := sheet.Columns[column]
	if !ok {
		return -1
	}
	return index
}

func Export(command string, sheetID string, entrySets []*EntrySet) (wg *sync.WaitGroup) {
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
			//"Generated by github.com/bleeding182/localization v" + version,
			"Do _not_ modify",
			"https://docs.google.com/spreadsheets/d/" + sheetID + "#gid=" + sheets[i].GID,
			fmt.Sprintf("Last updated at %v", timestamp),
		}

		createSheetModel(command, sheets[i])
	}

	for _, sheet := range sheets {
		wg.Add(1)
		go feedWriter(Writers[command], sheet, wg)
	}
	return
}

func parseEntrySetToSheet(entrySet *EntrySet, sheetChan chan *sheet) {
	fmt.Sprintln("Sheet", entrySet.GID)

	var sheet = &sheet{
		GID:     entrySet.GID,
		Locale:  entrySet.Locale,
		Columns: make(map[string]int),
		Data:    make([]writer.LocalizedString, 0, len(entrySet.Values)),
		Plurals: make(map[string]writer.QuantityString),
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

		compositeKey := writer.CompositeKeyOf(key)

		s := writer.LocalizedString{
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
				plural = writer.QuantityString{
					Key:    key,
					Values: make(map[writer.Quantity]writer.LocalizedString),
				}
				sheet.Plurals[key] = plural
			}

			plural.Values[writer.QuantityOf(compositeKey.Quantity())] = s
		}
	}

	sort.Slice(sheet.Data, func(i, j int) bool {
		return sheet.Data[i].Key.Original() < sheet.Data[j].Key.Original()
	})

	sheetChan <- sheet
}

func parse(row []interface{}, index int) string {
	if index >= 0 && index < len(row) {
		return row[index].(string)
	}
	return ""
}

func createSheetModel(tag string, sheet *sheet) {

	overrideIndex, ok := sheet.Columns[tag]
	if !ok {
		overrideIndex = -1
	}

	var w = Writers[tag]

	model := &writer.LocalizationModel{
		Headers: sheet.Headers,
		Groups:  make([]writer.Group, 0),
		Plurals: &sheet.Plurals,
	}

	var group writer.Group
	for _, ls := range sheet.Data {
		if group.Name != ls.Key.Group() {
			if group.Name != "" {
				model.Groups = append(model.Groups, group)
			}
			group = writer.Group{
				Name:    ls.Key.Group(),
				Strings: make([]writer.AndroidString, 0),
			}
		}
		var value string
		if overrideIndex >= 0 && overrideIndex < len(ls.Entries) && ls.Entries[overrideIndex] != "" {
			value = ls.Entries[overrideIndex]
		} else {
			value = html.EscapeString(w.Normalize(ls.Value))
		}

		group.Strings = append(group.Strings, writer.AndroidString{ls.Key.Original(), value, ls.Comment})
	}
	if group.Name != "" {
		model.Groups = append(model.Groups, group)
	}

	sheet.Model = model
}

func feedWriter(w writer.Writer, sheet *sheet, wg *sync.WaitGroup) {
	defer wg.Done()
	w.Export(sheet.Locale, sheet.Model)
}
