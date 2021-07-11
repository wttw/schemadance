package {{ .Package }}

import (
	"github.com/jackc/pgx/v4"
	"{{ .MigratePath }}/check"
	"{{ .MigratePath }}/check/checkpgx"
	"testing"
)

var ScratchDatabaseUrl = "postgresql://localhost/scratch_db"
var CreateDbUrl = "postgresqql://localhost/postgres"

var _ check.Checker = Checker{}

type Checker struct {
	db *pgx.Conn
}

func (c Checker) Fingerprint() (string, error) {
	return checkpgx.Fingerprint(c.db)
}


func TestMigrations(t *testing.T) {
	db, err := checkpgx.CreateScratchDb(ScratchDatabaseUrl, CreateDbUrl)
	if err != nil {
		t.Fatalf("failed to acquire scratch database %s: %v", ScratchDatabaseUrl, err)
	}
	check.CheckMigrations(t, NewMigrator(db), Checker{db})
}
