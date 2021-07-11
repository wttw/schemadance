package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Patch represents a sql script and associated Go functions to
// go from one schema version to another
type Patch struct {
	Prefix       string
	From         int
	To           int
	Description  string
	SqlFilename  string
	PreFunction  string
	PostFunction string
}

// Patches reads the filenames in the current directory and
// returns all the patch files for a given prefix
func Patches() (map[string][]Patch, map[string][]Patch, error) {
	migrationSeen := map[string]string{}

	upRe := regexp.MustCompile(`^([a-z]+)_(\d+)(?:\..*)?\.up\.sql$`)
	downRe := regexp.MustCompile(`^([a-z]+)_(\d+)(?:\..*)?\.down\.sql$`)
	toFromRe := regexp.MustCompile(`^([a-z]+)_(\d+)-(\d+)(?:\..*)?.sql$`)

	files, err := os.ReadDir(".")
	if err != nil {
		return nil, nil, err
	}

	up := map[string][]Patch{}
	down := map[string][]Patch{}

	for _, f := range files {
		var to, from int
		var prefix string
		if matches := upRe.FindStringSubmatch(f.Name()); matches != nil {
			prefix = matches[1]
			to, err = strconv.Atoi(matches[2])
			if err != nil {
				return nil, nil, err
			}
			if to < 1 {
				return nil, nil, fmt.Errorf("invalid destination version in %s", f.Name())
			}
			from = to - 1
		} else if matches := downRe.FindStringSubmatch(f.Name()); matches != nil {
			prefix = matches[1]
			from, err = strconv.Atoi(matches[2])
			if err != nil {
				return nil, nil, err
			}
			if from < 1 {
				return nil, nil, fmt.Errorf("invalid source version in %s", f.Name())
			}
			to = from - 1
		} else if matches := toFromRe.FindStringSubmatch(f.Name()); matches != nil {
			prefix = matches[1]
			from, err = strconv.Atoi(matches[2])
			if err != nil {
				return nil, nil, err
			}
			if from < 0 {
				return nil, nil, fmt.Errorf("invalid source version in %s", f.Name())
			}
			to, err = strconv.Atoi(matches[3])
			if err != nil {
				return nil, nil, err
			}
			if to < 0 {
				return nil, nil, fmt.Errorf("invalid destination version in %s", f.Name())
			}
		} else {
			continue
		}
		if to == from {
			log.Fatalf("%s tries to migrate from %d to %d", f.Name(), from, from)
		}
		key := fmt.Sprintf("%s-%d-%d", prefix, from, to)
		seen, ok := migrationSeen[key]
		if ok {
			log.Fatalf("%s and %s both migrate from version %d to %d", seen, f.Name(), from, to)
		}
		migrationSeen[key] = f.Name()

		if to > from {
			up[prefix] = append(up[prefix], Patch{
				Prefix:      prefix,
				From:        from,
				To:          to,
				SqlFilename: f.Name(),
			})
		} else {
			down[prefix] = append(down[prefix], Patch{
				Prefix:      prefix,
				From:        from,
				To:          to,
				SqlFilename: f.Name(),
			})
		}
	}
	return up, down, nil
}


// Read the SQL patch files to get any additional metadata
func populateMeta(patches []Patch) {
	descRe := regexp.MustCompile(`^--\s*Description:\s*(.*)`)
	beforeRe := regexp.MustCompile(`^--\s*Before:\s*(.*)`)
	afterRe := regexp.MustCompile(`^--\s*After:\s*(.*)`)
	for i, patch := range patches {
		file, err := os.Open(patch.SqlFilename)
		if err != nil {
			log.Fatalf("Failed to open %s: %v", patch.SqlFilename, err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "-- ") {
				break
			}
			if matches := descRe.FindStringSubmatch(line); matches != nil {
				patches[i].Description = matches[1]
			}
			if matches := beforeRe.FindStringSubmatch(line); matches != nil {
				patches[i].PreFunction = matches[1]
			}
			if matches := afterRe.FindStringSubmatch(line); matches != nil {
				patches[i].PostFunction = matches[1]
			}
		}
		if err := scanner.Err(); err != nil {
			log.Fatalln(err)
		}
		if err = file.Close(); err != nil {
			log.Fatalln(err)
		}
	}
}

// AddPatch creates an empty SQL file for up and downgrades
func AddPatch(desc string, prefix string) {
	upPatches, _, err := Patches()
	if err != nil {
		log.Fatalf("Failed to read existing patches: %v", err)
	}
	patches := upPatches[prefix]
	nextVersion := 1
	for _, p := range patches {
		if p.To >= nextVersion {
			nextVersion = p.To + 1
		}
	}
	slug := strings.Trim(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(desc, "_"), "_")
	if slug != "" {
		slug = slug + "."
	}

	makeSQLFile(fmt.Sprintf("%s_%03d.%sup.sql", prefix, nextVersion, slug), desc, nextVersion-1, nextVersion)
	makeSQLFile(fmt.Sprintf("%s_%03d.%sdown.sql", prefix, nextVersion, slug), desc, nextVersion, nextVersion-1)
}

func makeSQLFile(name string, description string, from int, to int) {
	tpl, err := Templates()
	if err != nil {
		log.Fatalln(err)
	}

	file, err := os.Create(name)

	if err != nil {
		log.Fatalf("Failed to create %s: %v", name, err)
	}
	err = tpl.ExecuteTemplate(file, "emptysql.tpl", struct {
		Description string
		From        int
		To          int
	}{
		Description: description,
		From:        from,
		To:          to,
	})
	if err != nil {
		log.Fatalf("Failed to create %s: %v", name, err)
	}
	err = file.Close()
	if err != nil {
		log.Fatalf("Failed to create %s: %v", name, err)
	}
}
