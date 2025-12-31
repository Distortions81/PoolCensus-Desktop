package main

import (
	"embed"
	"html/template"
)

//go:embed templates/*.tmpl
var presenterTemplates embed.FS

var tmpl = template.Must(loadTemplates())

func loadTemplates() (*template.Template, error) {
	funcMap := template.FuncMap{
		"fmtN": func(f float64, decimals int) string {
			return formatTrimmedFloat(f, decimals)
		},
		"fmtPct": func(f float64) string {
			return formatTrimmedFloat(f, 2) + "%"
		},
		"blockchainAddressURL": blockchainAddressURL,
	}

	t := template.New("poolcensus").Funcs(funcMap)
	files := []string{
		"templates/dashboard.tmpl",
		"templates/history.tmpl",
		"templates/details.tmpl",
	}
	return t.ParseFS(presenterTemplates, files...)
}
