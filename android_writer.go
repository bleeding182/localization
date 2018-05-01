package main

import (
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

const tagAndroid = "android"

func init() {
	writers[tagAndroid] = androidWriter{}
}

var regions *bool

func (writer androidWriter) registerCommand(app *kingpin.Application) {
	app.Command(tagAndroid, "Export your strings as xml for Android. All values from 'value' will be escaped, 'android' will be used as-is.\n\nPlurals can be added with a `__pl_<one|other|...>` suffix")
}

type androidWriter struct {
}

func (Writer androidWriter) Tag() string {
	return tagAndroid
}

func (Writer androidWriter) Export(locale string, model *LocalizationModel) {
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
	if err != nil {
		panic(err)
	}

	err = template.Execute(f, model)

	if err != nil {
		panic(err)
	}
}

var iosStringFormat = regexp.MustCompile("%(\\d\\$)?@")

func (Writer androidWriter) Normalize(s string) string {
	return iosStringFormat.ReplaceAllStringFunc(s, func(s string) string {
		return strings.Replace(s, "@", "s", 1)
	})
}
