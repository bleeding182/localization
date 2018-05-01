package android

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/bleeding182/localization/writer"

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

const tagAndroid = "android"

var regions *bool
var outputFolder *string

func (writer AndroidWriter) RegisterCommand(app *kingpin.Application) {
	command := app.Command(tagAndroid, "Export your strings as xml for Android. All values from 'value' will be escaped, 'android' will be used as-is.\n\nPlurals can be added with a `__pl_<one|other|...>` suffix")
	outputFolder = command.Flag("outputFolder", "Set the output directory where the values-* folders will be generated.").Default("exports").String()
}

type AndroidWriter struct{}

func (Writer AndroidWriter) Tag() string {
	return tagAndroid
}

func (Writer AndroidWriter) Export(locale string, model *writer.LocalizationModel) {
	var folder string
	if locale == "default" {
		folder = "values"
	} else {
		folder = "values-" + locale
	}

	localeFolder := path.Join(*outputFolder, folder)
	f := openFile(localeFolder, "generated_strings.xml")
	defer f.Close()

	template, err := template.New("file").Parse(androidTemplate)
	check(err)

	err = template.Execute(f, model)
	check(err)
}

var iosStringFormat = regexp.MustCompile("%(\\d\\$)?@")

func (Writer AndroidWriter) Normalize(s string) string {
	return iosStringFormat.ReplaceAllStringFunc(s, func(s string) string {
		return strings.Replace(s, "@", "s", 1)
	})
}

func openFile(folder string, name string) *os.File {
	foldername := fmt.Sprintf("%v", folder)
	os.MkdirAll(foldername, os.ModePerm)

	filename := fmt.Sprintf("%v/%v", foldername, name)
	f, err := os.Create(filename)
	check(err)
	return f
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
