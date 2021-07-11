package check

import (
	"context"
	"github.com/nsf/jsondiff"
	"github.com/wttw/schemadance"
	"sort"
	"testing"
)

type Checker interface {
	Fingerprint() (string, error)
	Version(set string) (int, error)
}

type MigrationPath struct {
	Steps []int
	Fingerprint string
}

func CheckMigrations(t *testing.T, migrator schemadance.Migrator, checker Checker) {
	ctx := context.Background()
	err := migrator.Database.Initialize(ctx)
	if err != nil {
		t.Fatalf("failed to initialize schema version: %v", err)
		return
	}

	// Get the names of each patch stream, so we can run each
	// one in a subtest.
	sets := []string{}
	for _, v := range migrator.Patches.Sets {
		sets = append(sets, v.Name)
	}
	sort.Strings(sets)

	for _, set := range sets {
		t.Run(set, func(t *testing.T){
			tx, err := migrator.Database.Begin(ctx)
			if err != nil {
				t.Fatalf("begin: %v", err)
				return
			}
			baseVersion, err := tx.Version(ctx, set)
			if err != nil {
				t.Fatalf("failed to get version: %v", err)
				return
			}
			tx.Commit(ctx)
			if baseVersion != 0 {
				t.Errorf("expected based version to be 0, not %d", baseVersion)
				return
			}
			for _, s := range migrator.Patches.Sets {
				if s.Name == set {
					fingerprints := map[int]MigrationPath{}
					fp, err := checker.Fingerprint()
					if err != nil {
						t.Fatalf("failed to fingerprint: %v", err)
						return
					}
					path := []int{0}
					fingerprints[baseVersion] = MigrationPath{
						Steps:       path,
						Fingerprint: fp,
					}

					maxVersion := s.EndVersion()

					// Up, one by one
					for to := 1; to <= maxVersion; to++ {
						path = append(path, to)
						beforeVersion := getVersion(t, checker, set)
						err := migrator.Migrate(ctx, set, to)
						if err != nil {
							t.Errorf("failed to migrate from %d to %d: %v", beforeVersion, to, err)
							return
						}
						afterVersion := getVersion(t, checker, set)
						if afterVersion != to {
							t.Errorf("Expected version to be %d, not %d after %d->%d migration", to, afterVersion, beforeVersion, to)
							return
						}
						if !checkFingerprint(t, fingerprints, checker, path, set) {
							return
						}
					}

					// Down, one by one
					for to := maxVersion; to >= 0; to-- {
						beforeVersion := getVersion(t, checker, set)
						err := migrator.Migrate(ctx, set, to)
						if err != nil {
							t.Errorf("failed to migrate from %d to %d: %v", beforeVersion, to, err)
							return
						}
						afterVersion := getVersion(t, checker, set)
						if afterVersion != to {
							t.Errorf("Expected version to be %d, not %d after %d->%d migration", to, afterVersion, beforeVersion, to)
							return
						}
						if !checkFingerprint(t, fingerprints, checker, path, set) {
							return
						}
					}
				}
			}
		})
	}
}

func getVersion(t *testing.T, checker Checker, set string) int {
	t.Helper()
	ver, err := checker.Version(set)
	if err != nil {
		t.Fatalf("failed to get version for %s: %v", set, err)
	}
	return ver
}

func checkFingerprint(t *testing.T, fingerprints map[int]MigrationPath, checker Checker, path []int, set string) bool {
	t.Helper()
	version := getVersion(t, checker, set)
	fp, err := checker.Fingerprint()
	if err != nil {
		t.Fatalf("failed to get fingerprint: %v", err)
	}
	existing, ok := fingerprints[version]
	if !ok {
		fingerprints[version] = MigrationPath{
			Steps:       path,
			Fingerprint: fp,
		}
		return true
	}
	if existing.Fingerprint == fp {
		return true
	}
	opts := jsondiff.DefaultConsoleOptions()
	diff, msg := jsondiff.Compare([]byte(existing.Fingerprint), []byte(fp), &opts)
	if diff == jsondiff.FullMatch {
		return true
	}
	t.Errorf("fingerprint mismatch between %v and %v: %s", existing.Steps, path, msg)
	return false
}
