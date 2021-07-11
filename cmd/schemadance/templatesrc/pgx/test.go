package dbpgx

import (
	"context"
	"github.com/jackc/pgx/v4"
	"github.com/wttw/schemadance/check"
	"github.com/wttw/schemadance/check/checkpgx"
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

func (c Checker) Version(prefix string) (int, error) {
	var version int
	err := c.db.QueryRow(context.Background(), `select version from schema_version where name = $1`, prefix).Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

func TestMigrations(t *testing.T) {
	db, err := checkpgx.CreateScratchDb(ScratchDatabaseUrl, CreateDbUrl)
	if err != nil {
		t.Fatalf("failed to acquire scratch database %s: %v", ScratchDatabaseUrl, err)
	}
	check.CheckMigrations(t, NewMigrator(db), Checker{db})
}
