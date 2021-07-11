package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// InitializeDir initializes a directory to hold schema patches
func InitializeDir(databaseType string, migratePackage string) error {
	if !ValidDbType(databaseType) {
		log.Fatalf("%q is not a supported database type\nvalid types are %s\n", databaseType, strings.Join(DbTypes(), ", "))
	}
	dbFile := "db.go"

	dir, _ := os.Getwd()

	base := filepath.Base(dir)

	for _, filename := range []string{base + ".go", dbFile, base + "_test.go"} {
		if _, err := os.Stat(filename); !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%s exists in %s, cannot init", filename, dir)
		}
	}

	f, err := os.Create(base + ".go")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(f, "package %s\n//go:generate schemadance --generate %s\n", base, databaseType)
	if err != nil {
		return err
	}

	data := struct {
		Package        string
		MigratePath    string
		MigratePackage string
	}{
		Package:        base,
		MigratePath:    migratePackage,
		MigratePackage: path.Base(migratePackage),
	}

	err = RenderGoTemplate("db_" + databaseType +".tpl", dbFile, data)
	if err != nil {
		return err
	}

	err = RenderGoTemplate("test_" + databaseType +".tpl", base + "_test.go", data)
	if err != nil {
		return err
	}

	return nil
}


