// Package trdsql implements execute SQL queries on tabular data.
//
// trdsql imports tabular data into a database,
// executes SQL queries, and executes exports.
package trdsql

import (
	"context"
	"fmt"
	"log"
)

// AppName is used for command names.
var AppName = "trdsql"

// TRDSQL represents DB definition and Importer/Exporter interface.
type TRDSQL struct {
	// Driver is database driver name(sqlite3/mysql/postgres).
	Driver string
	// Dsn is data source name.
	Dsn string

	// Importer is interface of processing to
	// import(create/insert) data.
	Importer Importer
	// Exporter is interface export to the process of
	//  export(select) from the database.
	Exporter Exporter
}

// NewTRDSQL returns a new TRDSQL structure.
func NewTRDSQL(im Importer, ex Exporter) *TRDSQL {
	return &TRDSQL{
		Driver:   "sqlite3",
		Dsn:      "",
		Importer: im,
		Exporter: ex,
	}
}

// Format represents the import/export format.
type Format int

// Represents Format.
const (
	// import (guesses for import format)
	GUESS Format = iota

	// import/export
	// Format using go standard CSV library.
	CSV

	// import/export
	// Labeled Tab-separated Values.
	LTSV

	// import/export
	// Format using go standard JSON library.
	JSON

	// import/export
	// TBLN format(https://tbln.dev).
	TBLN

	// export
	// Output as it is.
	// Multiple characters can be selected as delimiter.
	RAW

	// export
	// MarkDown format.
	MD

	// export
	// ASCII Table format.
	AT

	// export
	// Vertical format.
	VF

	// export
	// JSON Lines format(http://jsonlines.org/).
	JSONL
)

func (f Format) String() string {
	switch f {
	case GUESS:
		return "GUESS"
	case CSV:
		return "CSV"
	case LTSV:
		return "LTSV"
	case JSON:
		return "JSON"
	case TBLN:
		return "TBLN"
	case RAW:
		return "RAW"
	case MD:
		return "MD"
	case AT:
		return "AT"
	case VF:
		return "VF"
	case JSONL:
		return "JSONL"
	default:
		return "Unknown"
	}
}

// Exec is actually executed.
func (trd *TRDSQL) Exec(sql string) error {
	ctx := context.Background()
	return trd.ExecContext(ctx, sql)
}

// Exec is actually executed.
func (trd *TRDSQL) ExecContext(ctx context.Context, sql string) error {
	db, err := Connect(trd.Driver, trd.Dsn)
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	defer func() {
		err = db.Disconnect()
		if err != nil {
			log.Printf("disconnect: %s", err)
		}
	}()

	db.Tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	if trd.Importer != nil {
		sql, err = trd.Importer.ImportContext(ctx, db, sql)
		if err != nil {
			return fmt.Errorf("import: %w", err)
		}
	}

	if trd.Exporter != nil {
		err = trd.Exporter.ExportContext(ctx, db, sql)
		if err != nil {
			return fmt.Errorf("export: %w", err)
		}
	}

	err = db.Tx.Commit()
	if err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	return nil
}
