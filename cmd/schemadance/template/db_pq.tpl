package {{ .Package }}

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v4"
	"{{ .MigratePath }}"
)

var _ {{ .MigratePackage }}.Tx = Tx{}
var _ {{ .MigratePackage }}.Db = Db{}

func NewMigrator(db *sql.Conn) {{ .MigratePackage }}.Migrator {
    return {{ .MigratePackage }}.Migrator{
        Database: Db{Conn: db},
        Patches: PatchSets,
    }
}


type Tx struct {
	*sql.Tx
}

func (t Tx) Version(ctx context.Context, prefix string) (int, error) {
	var version int
	err := t.Tx.QueryRowContext(ctx, `select version from schema_version where name = $1`, prefix).Scan(&version)
	if err == nil {
		return version, nil
	}
	if err != pgx.ErrNoRows {
		return 0, err
	}
	_, err = t.Tx.ExecContext(ctx, `insert into schema_version (name, version) values ($1, 0)`, prefix)
	return 0, err
}

func (t Tx) SetVersion(ctx context.Context, prefix string, version int) error {
	tag, err := t.Tx.ExecContext(ctx, `update schema_version set version = $2 where name = $1`, prefix, version)
	if err != nil {
		return err
	}
	rowsAffected, err := tag.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return fmt.Errorf("%d rows updated in schema_version for %s, expected 1", rowsAffected, prefix)
	}
	return nil
}

func (t Tx) Exec(ctx context.Context, sql string) error {
	_, err := t.Tx.ExecContext(ctx, sql)
	return err
}

func (t Tx) Rollback(_ context.Context) error {
	return t.Tx.Rollback()
}

func (t Tx) Commit(_ context.Context) error {
	return t.Tx.Commit()
}

type Db struct {
	*sql.Conn
}

func (d Db) Initialize(ctx context.Context) error {
	tx, err := d.Begin(ctx)
	if err != nil {
		return err
	}
	// language=SQL
	err = tx.Exec(ctx, `create table if not exists schema_version (name primary key, version int)`)
	if err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

func (d Db) Begin(ctx context.Context) ({{ .MigratePackage }}.Tx, error) {
	tx, err := d.Conn.BeginTx(ctx, nil)
	return Tx{tx}, err
}

func (d Db) Close(_ context.Context) error {
	return d.Conn.Close()
}


