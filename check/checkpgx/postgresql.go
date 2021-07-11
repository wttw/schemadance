package checkpgx

//go:generate go run ../generate/loadstrings.go pgx Column Sequence Table Function

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/jackc/pgx/v4"
	"strconv"
)


type Column struct {
	TableSchema string
	TableName string
	ColumnName string
	ColumnDefault string
	IsNullable string
	DataType string
	CharacterMaximumLength string
	CharacterOctetLength string
	NumericPrecision string
	NumericPrecisionRadix string
	DatetimePrecision string
	IntervalType string
	IntervalPrecision string
}

type Sequence struct {
	SequenceSchema string
	SequenceName string
}

type Table struct {
	TableSchema string
	TableName string
	TableType string
	IsInsertableInto string
	Columns []Column `load:"-"`
}

type Function struct {
	RoutineSchema string
	RoutineName string
	RoutineType string
	RoutineBody string
	RoutineDefinition string
	IsDeterministic string
	IsNullCall string
	SecurityType string
}

type Everything struct {
	Tables map[string]Table
	Functions []Function
	Sequences []Sequence
}

// Convert a, possibly quoted, {schema, name} pair to an id string
func schemaName(schema, name string) string {
	s, err := strconv.Unquote(schema)
	if err != nil {
		s = schema
	}
	n, err := strconv.Unquote(name)
	if err != nil {
		n = name
	}
	return s + "." + n
}

// Fingerprint grovels through a database and returns a, probably JSON,
// string that will be the same if the database schema is roughly
// the same. We use that to detect common errors in migration scripts.
func Fingerprint(db *pgx.Conn) (string, error) {
	snap := Everything{
		Tables: map[string]Table{},
	}
	tables, err := LoadTable(db, `from information_schema.tables where table_schema not in ('pg_catalog', 'information_schema') order by table_schema, table_name`)
	if err != nil {
		return "", err
	}
	for _, t := range tables {
		snap.Tables[schemaName(t.TableSchema, t.TableName)] = t
	}

	columns, err := LoadColumn(db, `from information_schema.columns where table_schema not in ('pg_catalog', 'information_schema') order by table_schema, table_name, column_name`)
	if err != nil {
		return "", err
	}
	for _, c := range columns {
		cid := schemaName(c.TableSchema, c.TableName)
		t, ok := snap.Tables[cid]
		if !ok {
			return "", fmt.Errorf("column %s in %s has no matching table", cid, c.ColumnName)
		}
		t.Columns = append(t.Columns, c)
		snap.Tables[cid] = t
	}

	sequences, err := LoadSequence(db, `from information_schema.sequences order by sequence_schema, sequence_name`)
	if err != nil {
		return "", err
	}
	snap.Sequences = sequences

	functions, err := LoadFunction(db, `from information_schema.routines where routine_schema not in ('pg_catalog', 'information_schema') order by routine_schema, routine_type, routine_name, data_type`)
	if err != nil {
		return "", err
	}

	snap.Functions = functions

	var buff bytes.Buffer
	encoder := json.NewEncoder(&buff)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(snap)
	if err != nil {
		return "", err
	}
	return buff.String(), nil
}

// CreateScratchDb creates a new, clean database  for scratchDbUrl, using the
// superuser or user with createdb credentials given in createDbUrl
func CreateScratchDb(scratchDbUrl string, createDbUrl string) (*pgx.Conn, error) {
	ctx := context.Background()
	scratchCfg, err := pgx.ParseConfig(scratchDbUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot parse scratchDb url (%s): %w", scratchDbUrl, err)
	}

	cdbConn, err := pgx.Connect(ctx, createDbUrl)
	if err != nil {
		return nil, fmt.Errorf("cannot connect to config db (%s): %w", createDbUrl, err)
	}

	var versionString string
	err = cdbConn.QueryRow(ctx, `select current_setting('server_version_num')`).Scan(&versionString)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve postgresql version")
	}
	version, err := strconv.Atoi(versionString)
	if err != nil {
		return nil, fmt.Errorf("invalid version returned (%s): %w", versionString, err)
	}
	if version >= 130000 {
		// Postgresql >= 13.0 has "drop table force"
		_, err = cdbConn.Exec(ctx, fmt.Sprintf(`drop database if exists %s with (force)`), scratchCfg.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to drop database %s: %w", scratchCfg.Database, err)
		}
	} else {
		// Postgresql < 13.0; disconnect everyone explicitly. A bit racier than using force.
		_, err = cdbConn.Exec(ctx, `SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE datname = $1`, scratchCfg.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to disconnect users: %w", err)
		}
		_, err = cdbConn.Exec(ctx, fmt.Sprintf(`drop database if exists %s`), scratchCfg.Database)
		if err != nil {
			return nil, fmt.Errorf("failed to drop database %s: %w", scratchCfg.Database, err)
		}
	}

	_, err = cdbConn.Exec(ctx, fmt.Sprintf(`create database %s owner %s`, scratchCfg.Database, scratchCfg.User))
	if err != nil {
		return nil, fmt.Errorf("failed to create scratch database (%s): %w", scratchDbUrl, err)
	}
	err = cdbConn.Close(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to close admin database connection: %w", err)
	}
	conn, err := pgx.ConnectConfig(ctx, scratchCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to scratch database (%s): %w", scratchDbUrl, err)
	}
	return conn, nil
}