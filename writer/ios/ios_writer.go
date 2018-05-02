package ios

import (
	"fmt"
	"os"
	"path"
	"regexp"
	"strings"
	"text/template"

	"github.com/bleeding182/localization/writer"
	"github.com/iancoleman/strcase"

	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

const iosStringsTemplate = `
{{range $header := $.Headers -}}
/* {{$header}} */
{{end -}}

{{- range $g := $.Groups}}
/** {{.Name}} **/
{{- range $as := $g.Strings}}
{{- if .Comment}}
/* {{.Comment}} */
{{- end}}
"{{.Key}}" = "{{.Value}}";
{{- end}}
{{end}}
`

const iosStringsDictTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
{{range $header := $.Headers -}}
<!-- {{$header}} -->
{{end -}}
<plist version="1.0">
{{- range $p := $.Plurals}}
    <dict>
        <key>{{.Key}}</key>
        <dict>
            <key>NSStringLocalizedFormatKey</key>
            <string>%#@{{.Key}}Key@</string>
            <key>{{.Key}}Key</key>
            <dict>
                <key>NSStringFormatSpecTypeKey</key>
                <string>NSStringPluralRuleType</string>
                <key>NSStringFormatValueTypeKey</key>
                <string>d</string>
            {{- range $q, $v := .Values}}
                <key>{{$q}}</key>
                <string>{{$v.Value}}</string>
            {{- end}}
            </dict>
        </dict>
    </dict>
{{end}}
</plist>
`

const iosStringsUtilTemplate = `
import Foundation

{{range $header := $.Headers -}}
// {{$header}}
{{end}}
// swiftlint:disable line_length
public struct Strings {

    {{- range $g := $.Groups}}
    public struct {{.Name | camelcase}} {
        {{- range $as := $g.Strings}}
        static let {{.Key | camelcase}} = Strings.localized("{{.Key}}", value: "{{.Value}}"
            {{- if .Comment -}}
                , comment: "{{.Comment}}"
            {{- end -}}
        )
        {{- end}}
    }
    {{end}}

    public static func localized(_ key: String, tableName: String? = nil, bundle: Bundle = Bundle.main, value: String, comment: String = "") -> String {
        return NSLocalizedString(key, tableName: tableName, bundle: bundle, value: value, comment: comment)
    }
}
`

const tagIos = "ios"

var stringsFolder, utilFolder *string

func (writer IOSWriter) RegisterCommand(app *kingpin.Application) {
	command := app.Command(tagIos, "Export your strings for iOS. This will generate LocalizableGen.strings in *.lproj folders along with a Strings.swift util class.")
	stringsFolder = command.Flag("outputFolder", "Set the output directory where the *.lproj folders will be generated.").Default("exports").String()
	utilFolder = command.Flag("utilOutputFolder", "Set the output directory where the util-file will be generated. This Strings.swift file contains constants for easier access.").Default("exports").String()
}

type IOSWriter struct{}

func (writer IOSWriter) Tag() string {
	return tagIos
}

var funcs = template.FuncMap{
	"camelcase": strcase.ToCamel,
}

func (writer IOSWriter) Export(locale string, model *writer.LocalizationModel) {
	var folder string
	if locale == "default" {
		folder = "Base.lproj"
	} else {
		folder = locale + ".lproj"
	}

	localeFolder := path.Join(*stringsFolder, folder)

	stringsTemplate, err := template.New("strings").Parse(iosStringsTemplate)
	check(err)
	pluralsTemplate, err := template.New("plurals").Parse(iosStringsDictTemplate)
	check(err)

	stringsFile := openFile(localeFolder, "LocalizableGen.strings")
	defer stringsFile.Close()
	err = stringsTemplate.Execute(stringsFile, model)
	check(err)

	stringsDictFile := openFile(localeFolder, "LocalizableGen.stringsdict")
	defer stringsDictFile.Close()
	err = pluralsTemplate.Execute(stringsDictFile, model)
	check(err)

	if locale == "default" {
		utilTemplate, err := template.New("util").Funcs(funcs).Parse(iosStringsUtilTemplate)
		check(err)
		stringsUtilFile := openFile(*utilFolder, "Strings.swift")
		defer stringsUtilFile.Close()
		err = utilTemplate.Execute(stringsUtilFile, model)
		check(err)
	}
}

var androidStringFormat = regexp.MustCompile("%(\\d\\$)?s")

func (writer IOSWriter) Normalize(s string) string {
	var formatted = androidStringFormat.ReplaceAllStringFunc(s, func(s string) string {
		return strings.Replace(s, "s", "@", 1)
	})
	return strings.Replace(formatted, "\"", "\\\"", -1)
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
