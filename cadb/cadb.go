// Package cadb manages the persistent state necessary for the mini CA
// that liteboot acts as.

package cadb // import "github.com/Linaro/lite_bootstrap_server/cadb"

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type Conn struct {
	db *sql.DB
}

func Open() (*Conn, error) {
	db, err := sql.Open("sqlite3", "CADB.db")
	if err != nil {
		return nil, err
	}

	conn := &Conn{
		db: db,
	}

	err = conn.checkSchema()
	if err != nil {
		fmt.Printf("database error: %v\n", err)
		return nil, err
	}

	return conn, nil
}
