// Code generated by schemadance DO NOT EDIT.
package {{.Package}}

import (
    "{{.MigratePath}}"
)

var {{ title .Prefix}} = {{.MigratePackage}}.PatchSet {
    Name: {{ printf "%q" .Prefix }},
    Up: []{{.MigratePackage}}.Patch{
        {{ range .Up }}{
            Prefix: {{ printf "%q" .Prefix }},
            From: {{.From}},
            To: {{.To}},
            Description: {{ printf "%q" .Description}},
            Filename: {{ printf "%q" .SqlFilename}},
            Before: {{ with .PreFunction }}{{.}}{{else}}nil{{end}},
            After: {{ with .PostFunction }}{{.}}{{else}}nil{{end}},
        },
{{end}}
    },
    Down: []{{.MigratePackage}}.Patch{
        {{ range .Down }}{
            Prefix: {{ printf "%q" .Prefix }},
            From: {{.From}},
            To: {{.To}},
            Description: {{ printf "%q" .Description}},
            Filename: {{ printf "%q" .SqlFilename}},
            Before: {{ with .PreFunction }}{{.}}{{else}}nil{{end}},
            After: {{ with .PostFunction }}{{.}}{{else}}nil{{end}},
        },
{{end}}
    },
}