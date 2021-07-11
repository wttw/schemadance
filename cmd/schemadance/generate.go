package main

import (
	"log"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
)

// GenerateFiles reads in the sql patches in the current directory
// and generates .go files that contain their metadata and embed
// the sql.
func GenerateFiles(databaseType string, migratePackage string) {
	if !ValidDbType(databaseType) {
		log.Fatalf("%q is not a supported database type\nvalid types are %s\n", databaseType, strings.Join(DbTypes(), ", "))
	}
	dir, _ := os.Getwd()
	base := filepath.Base(dir)

	ups, downs, err := Patches()
	if err != nil {
		log.Fatalln(err)
	}

	var allFiles []string

	patchSets := make([]string, 0, len(ups))
	for name := range ups {
		patchSets = append(patchSets, name)
	}
	sort.Strings(patchSets)

	for _, prefix := range patchSets {
		up := ups[prefix]
		down := downs[prefix]
		sort.Slice(up, func(i, j int) bool {
			if up[i].From == up[j].From {
				return up[i].To < up[j].To
			}
			return up[i].From < up[j].From
		})
		sort.Slice(down, func(i, j int) bool {
			if down[i].From == down[j].From {
				return down[i].To > down[j].To
			}
			return down[i].From > down[j].From
		})

		populateMeta(down)
		populateMeta(up)

		files := ""
		for _, p := range up {
			if len(files) > 0 && len(files)+len(p.SqlFilename) > 65 {
				allFiles = append(allFiles, files)
				files = ""
			}
			files = files + " " + p.SqlFilename
		}
		allFiles = append(allFiles, files)

		files = ""
		for _, p := range down {
			if len(files) > 0 && len(files)+len(p.SqlFilename) > 65 {
				allFiles = append(allFiles, files)
				files = ""
			}
			files = files + " " + p.SqlFilename
		}
		allFiles = append(allFiles, files)

		err := RenderGoTemplate("patches.tpl", prefix+".gen.go", struct {
			Package        string
			MigratePath    string
			MigratePackage string
			Prefix         string
			Up             []Patch
			Down           []Patch
		}{
			Package:        base,
			MigratePath:    migratePackage,
			MigratePackage: path.Base(migratePackage),
			Prefix:         prefix,
			Up:             up,
			Down:           down,
		})
		if err != nil {
			log.Fatalln(err)
		}
	}

	err = RenderGoTemplate("patch_set.tpl", "patch_set.gen.go", struct {
		Package        string
		MigratePath    string
		MigratePackage string
		Sets           []string
		Files          []string
	}{
		Package:        base,
		MigratePath:    migratePackage,
		MigratePackage: path.Base(migratePackage),
		Sets:           patchSets,
		Files:          allFiles,
	})
	if err != nil {
		log.Fatalln(err)
	}
}
