package dbpgx

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4"
	"github.com/wttw/schemadance"
)

var _ schemadance.Tx = Tx{}
var _ schemadance.Db = Db{}

func NewMigrator(db *pgx.Conn) schemadance.Migrator {
    return schemadance.Migrator{
        Database: Db{Conn: db},
        Patches: PatchSets,
    }
}


type Tx struct {
  pgx.Tx
}

func (t Tx) Version(ctx context.Context, prefix string) (int, error) {
	var version int
	err := t.Tx.QueryRow(ctx, `select version from schema_version where name = $1`, prefix).Scan(&version)
	if err == nil {
		return version, nil
	}
	if err != pgx.ErrNoRows {
		return 0, err
	}
	_, err = t.Tx.Exec(ctx, `insert into schema_version (name, version) values ($1, 0)`, prefix)
	return 0, err
}

func (t Tx) SetVersion(ctx context.Context, prefix string, version int) error {
	tag, err := t.Tx.Exec(ctx, `update schema_version set version = $2 where name = $1`, prefix, version)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("%d rows updated in schema_version for %s, expected 1", tag.RowsAffected(), prefix)
	}
	return nil
}

func (t Tx) Exec(ctx context.Context, sql string) error {
	_, err := t.Tx.Exec(ctx, sql)
	return err
}

func (t Tx) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}

func (t Tx) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

type Db struct {
	*pgx.Conn
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

func (d Db) Begin(ctx context.Context) (schemadance.Tx, error) {
	tx, err := d.Conn.Begin(ctx)
	return Tx{tx}, err
}

func (d Db) Close(ctx context.Context) error {
	return d.Close(ctx)
}
