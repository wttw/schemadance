// Code generated by generate.go DO NOT EDIT.

package checkpgx

import (
	"context"
	"github.com/jackc/pgx/v4"
)

func LoadSequence(db *pgx.Conn, sqlClause string) ([]Sequence, error) {
	sql := `select format("%L", sequence_schema, format("%L", sequence_name from ` + sqlClause
	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	var ret []Sequence
	for rows.Next() {
		var c Sequence
		err = rows.Scan(
			&c.SequenceSchema,
			&c.SequenceName,
		)
		if err != nil {
			return nil, err
		}
		ret = append(ret, c)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return ret, nil
}
