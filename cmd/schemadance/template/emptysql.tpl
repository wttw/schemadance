`-- {{if gt .To .From}}Upgrade{{else}}Downgrade{{end}} schema from version {{.From}} to {{.To}}
{{with .Description}}-- Description: {{ . }}
{{end}}
