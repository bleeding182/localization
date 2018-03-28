package main

import (
	"log"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// go build -ldflags "-X main.version={version}"
var version string

var (
	app = kingpin.New("localization", "A CLI-Tool to export your Google Sheets localizations for your project.").Version(version)
	// verbose = app.Flag("verbose", "Verbose logs. Use this to debug potential errors.").Bool()
	sheetID = app.Flag("sheetID", "ID of the spreadsheet to use.").Short('s').Required().String()
)

func main() {
	registerCommands(app)
	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	resp, err := loadSpreadSheet()

	if err != nil {
		log.Fatalf("Unable to retrieve data from sheet. %v", err)
	}

	entrySets := make([]*entrySet, len(resp))
	for i, data := range resp {
		sheet := data.sheet
		entrySets[i] = &entrySet{
			GID:     data.sheetID,
			Locale:  strings.Split(sheet.Range, "!")[0],
			Headers: sheet.Values[:1][0],
			Values:  sheet.Values[1:],
		}
	}

	wg := export(command, *sheetID, entrySets)

	wg.Wait()
}
