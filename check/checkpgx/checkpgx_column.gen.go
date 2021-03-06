// Code generated by generate.go DO NOT EDIT.

package checkpgx

import (
	"context"
	"github.com/jackc/pgx/v4"
)

func LoadColumn(db *pgx.Conn, sqlClause string) ([]Column, error) {
	sql := `select format("%L", table_schema, format("%L", table_name, format("%L", column_name, format("%L", column_default, format("%L", is_nullable, format("%L", data_type, format("%L", character_maximum_length, format("%L", character_octet_length, format("%L", numeric_precision, format("%L", numeric_precision_radix, format("%L", datetime_precision, format("%L", interval_type, format("%L", interval_precision from ` + sqlClause
	rows, err := db.Query(context.Background(), sql)
	if err != nil {
		return nil, err
	}
	var ret []Column
	for rows.Next() {
		var c Column
		err = rows.Scan(
			&c.TableSchema,
			&c.TableName,
			&c.ColumnName,
			&c.ColumnDefault,
			&c.IsNullable,
			&c.DataType,
			&c.CharacterMaximumLength,
			&c.CharacterOctetLength,
			&c.NumericPrecision,
			&c.NumericPrecisionRadix,
			&c.DatetimePrecision,
			&c.IntervalType,
			&c.IntervalPrecision,
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
