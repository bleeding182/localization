package main

import (
	"fmt"
	"html"
	"os"
	"path"
	"regexp"
	"strings"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const tag = "android"

func init() {
	writers[tag] = androidWriter{}
}

var regions *bool
var outputFolder *string

func (writer androidWriter) registerCommand(app *kingpin.Application) {
	command := app.Command(tag, "Export your strings as xml for Android. All values from 'value' will be escaped, 'android' will be used as-is.\n\nPlurals can be added with a `__pl_<one|other|...>` suffix")
	regions = command.Flag("regions", "Generate <!-- region group --> comments to group strings.").Bool()
	outputFolder = command.Flag("outputFolder", "Set the output directory where the values-* folders will be generated.").Default("exports").String()
}

type androidWriter struct {
}

func (writer androidWriter) Tag() string {
	return tag
}

func (writer androidWriter) Export(sheet *sheet) {
	var folder string
	if sheet.Locale == "default" {
		folder = "values"
	} else {
		folder = "values-" + sheet.Locale
	}

	localeFolder := path.Join(*outputFolder, folder)
	f := openFile(localeFolder, "generated_strings")
	defer f.Close()

	writeLine(f, "<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")

	for _, header := range sheet.Headers {
		writeLine(f, fmt.Sprintf("<!-- %v -->\n", header))
	}

	overrideIndex, ok := sheet.Columns[tag]
	if !ok {
		overrideIndex = -1
	}

	writeLine(f, "<resources>\n")
	defer writeLine(f, "</resources>\n")

	var group string

	for _, ls := range sheet.Data {
		if *regions && group != ls.Key.group() {
			if group != "" {
				writeLine(f, "    <!-- endregion -->\n\n")
			}
			writeLine(f, fmt.Sprintf("    <!-- region %v -->\n", ls.Key.group()))
			group = ls.Key.group()
		}
		if ls.Comment != "" {
			writeLine(f, fmt.Sprintf("    <!-- %v -->\n", ls.Comment))
		}

		var value string
		if overrideIndex >= 0 && overrideIndex < len(ls.Entries) && ls.Entries[overrideIndex] != "" {
			value = ls.Entries[overrideIndex]
		} else {
			value = html.EscapeString(writer.normalize(ls.Value))
		}

		writeLine(f, fmt.Sprintf("    <string name=\"%v\">%v</string>\n", ls.Key.original(), value))
	}
	if *regions && group != "" {
		writeLine(f, "    <!-- endregion -->\n\n")
	}

	for _, ps := range sheet.Plurals {
		writeLine(f, fmt.Sprintf("    <plurals name=\"%v\">\n", ps.Key))
		for key, item := range ps.Values {
			writeLine(f, fmt.Sprintf("        <item quantity=\"%v\">@string/%v</item>\n", key, item.Key.original()))
		}
		writeLine(f, "    </plurals>\n")
	}
}

var iosStringFormat = regexp.MustCompile("%(\\d\\$)?@")

func (writer androidWriter) normalize(s string) string {
	return iosStringFormat.ReplaceAllStringFunc(s, func(s string) string {
		return strings.Replace(s, "@", "s", 1)
	})
}

func writeLine(f *os.File, line string) {
	_, err := f.WriteString(line)
	check(err)
}

func openFile(folder string, name string) *os.File {
	foldername := fmt.Sprintf("%v", folder)
	os.MkdirAll(foldername, os.ModePerm)

	filename := fmt.Sprintf("%v/%v.xml", foldername, name)
	f, err := os.Create(filename)
	check(err)
	return f
}
