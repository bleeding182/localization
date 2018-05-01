package main

import (
	"path"
	"regexp"
	"strings"
	"text/template"

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

	// settings this closure allows you to use a custom localization provider, such as OneSky over-the-air
    // by default, NSLocalizedString will load the strings from the main bundle's Localizable.strings file
    public static var customLocalizationClosure: ((String, String?, Bundle, String, String) -> String)? = nil
    
    public static func localized(_ key: String, tableName: String? = nil, bundle: Bundle = Bundle.main, value: String, comment: String = "") -> String {
        if let closure = Strings.customLocalizationClosure {
            return closure(key, tableName, bundle, value, comment)
        } else {
            return NSLocalizedString(key, tableName: tableName, bundle: bundle, value: value, comment: comment)
        }
    }
}
`

const tagIos = "ios"

func init() {
	writers[tagIos] = iosWriter{}
}

func (writer iosWriter) registerCommand(app *kingpin.Application) {
	app.Command(tagIos, "")
}

type iosWriter struct {
}

func (writer iosWriter) Tag() string {
	return tagIos
}

var funcs = template.FuncMap{
	"camelcase": strcase.ToCamel,
}

func (writer iosWriter) Export(locale string, model *LocalizationModel) {
	var folder string
	if locale == "default" {
		folder = "values"
	} else {
		folder = "values-" + locale
	}

	localeFolder := path.Join(*outputFolder, folder)

	stringsTemplate, err := template.New("strings").Parse(iosStringsTemplate)
	if err != nil {
		panic(err)
	}
	utilTemplate, err := template.New("util").Funcs(funcs).Parse(iosStringsUtilTemplate)
	if err != nil {
		panic(err)
	}
	pluralsTemplate, err := template.New("plurals").Parse(iosStringsDictTemplate)
	if err != nil {
		panic(err)
	}

	stringsFile := openFile(localeFolder, "Localizable.strings")
	defer stringsFile.Close()
	err = stringsTemplate.Execute(stringsFile, model)
	if err != nil {
		panic(err)
	}

	stringsUtilFile := openFile(localeFolder, "Util.swift")
	defer stringsUtilFile.Close()
	err = utilTemplate.Execute(stringsUtilFile, model)
	if err != nil {
		panic(err)
	}

	stringsDictFile := openFile(localeFolder, "Localizable.stringsdict")
	defer stringsDictFile.Close()
	err = pluralsTemplate.Execute(stringsDictFile, model)
	if err != nil {
		panic(err)
	}

}

var androidStringFormat = regexp.MustCompile("%(\\d\\$)?s")

func (writer iosWriter) Normalize(s string) string {
	var formatted = iosStringFormat.ReplaceAllStringFunc(s, func(s string) string {
		return strings.Replace(s, "s", "@", 1)
	})
	return strings.Replace(formatted, "\"", "\\\"", -1)
}
