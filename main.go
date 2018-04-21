package main

import (
	"log"
	"os"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

// go build -ldflags "-X main.version={version}"
var version string

const appDescription = `
A CLI-Tool to export localized strings from your Google Sheets.

key             value         android comment
main_greeting	Hello, world!         Default greeting.

Keys:
    The keys used for your strings must match "<group>_<identifier>". If yor group name consists of multiple parts you can use "__" to mark it accordingly. Groups are used to group strings together.

Plurals:
	Plurals must be marked by the "__pl_<zero|one|two|few|many|other>" suffix on your key. If supported by the export target, they will be exported and grouped accordingly.
	
Values:
    Values may be escaped by the targte platform or modified in some other way. If you want to override a value you can place it in a column of the target platforms name.

For more information please also check https://github.com/bleeding182/localization where this app is published under the MIT Open Source License.
`

var (
	app = kingpin.New("localization", appDescription).Version(version)
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
