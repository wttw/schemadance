package main

import (
	flag "github.com/spf13/pflag"
	"log"
)

const defaultMigratePackage = "github.com/wttw/schemadance"

func main() {
	init := flag.String("init", "", "Initialize the current directory for schema patches")
	generate := flag.String("generate", "", "Generate code for schema migration")
	add := flag.String("add", "", "Add a new schema patch template with this description")
	prefix := flag.String("prefix", "patch", "Use this prefix for a new schema patch, default 'patch'")
	flag.Parse()

	migratePackage := defaultMigratePackage

	if *init != "" {
		err := InitializeDir(*init, migratePackage)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	if *generate != "" {
		GenerateFiles(*generate, migratePackage)
		return
	}

	if *add != "" {
		AddPatch(*add, *prefix)
	}
}
