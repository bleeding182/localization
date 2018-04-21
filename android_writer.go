package main

import (
	"fmt"
	"html"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const androidTemplate = `
<?xml version=\"1.0\" encoding=\"utf-8\"?>
{{range $header := $.Headers -}}
<!-- {{$header}} -->
{{end -}}
<resources>
{{- range $g := $.Groups}}
    <!-- region {{.Name}} -->
    {{- range $as := $g.Strings}}
    {{- if .Comment}}
    <!-- {{.Comment}} -->
    {{- end}}
    <string name="{{.Key}}">{{.Value}}</string>
    {{- end}}
    <!-- endregion -->
{{end}}
    <!-- region Plurals -->
{{- range $p := $.Plurals}}
    <plurals name="{{.Key}}">
    {{- range $q, $v := .Values}}
        <item quantity="{{$q}}">@string/{{$v.Key.Original}}</item>
    {{- end}}
    </plurals>
    <!-- endregion -->
{{end}}
</resources>
`

const tag = "android"

type AndroidModel struct {
	Headers []string
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

func init() {
	writers[tag] = androidWriter{}
}

var regions *bool
var outputFolder *string

func (writer androidWriter) registerCommand(app *kingpin.Application) {
	command := app.Command(tag, "Export your strings as xml for Android. All values from 'value' will be escaped, 'android' will be used as-is.\n\nPlurals can be added with a `__pl_<one|other|...>` suffix")
	outputFolder = command.Flag("outputFolder", "Set the output directory where the values-* folders will be generated.").Default("exports").String()
	//	regions = command.Flag("regions", "Generate <!-- region group --> comments to group strings.").Bool()
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

	overrideIndex, ok := sheet.Columns[tag]
	if !ok {
		overrideIndex = -1
	}

	model := &AndroidModel{
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
			value = html.EscapeString(writer.normalize(ls.Value))
		}

		group.Strings = append(group.Strings, AndroidString{ls.Key.Original(), value, ls.Comment})
	}
	if group.Name != "" {
		model.Groups = append(model.Groups, group)
	}

	template, err := template.New("file").Parse(androidTemplate)
	if err != nil {
		panic(err)
	}

	err = template.Execute(f, model)

	if err != nil {
		panic(err)
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
