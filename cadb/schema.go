package cadb

import "github.com/google/uuid"

// This is the schema for the database.  These statements will be
// evaluated in order to create the initial database.
var schema = []string{
	// settings holds key/value pairs for various settings related
	// to this CA.  There is a 'schemaVersion' setting that
	// ensures we have the correct database schema.
	`CREATE TABLE settings (key STRING PRIMARY KEY,
		value STRING NOT NULL)`,

	// devices holds all devices known to the system.  The id is
	// the identifier from the CSR.  Presumably, additional
	// information can be stored here.
	`CREATE TABLE devices (id STRING PRIMARY KEY)`,

	// certs holds all of the certificates we've ever issued.
	`CREATE TABLE certs (id STRING NOT NULL REFERENCES devices(id),
		serial STRING NOT NULL,
		cert BLOB NOT NULL,
		PRIMARY KEY (id, serial))`,
}

const schemaVersion = "20200820b"

func (conn *Conn) checkSchema() error {
	// Query the settings table for the schema version.
	row := conn.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, "schemaVersion")
	var version string
	err := row.Scan(&version)
	if err != nil {
		// Assume if there are no settings, then the database
		// is empty, and add this schema.
		return conn.setSchema()
	}

	if version == schemaVersion {
		return nil
	}

	panic("TODO: Implement database upgrades")
	return nil
}

// setSchema installs the above database schema into the connected
// database.
func (conn *Conn) setSchema() error {
	tx, err := conn.db.Begin()
	if err != nil {
		return err
	}

	for _, item := range schema {
		_, err = tx.Exec(item)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// Insert the schema version.
	_, err = tx.Exec(`INSERT INTO settings VALUES ('schemaVersion', ?)`, schemaVersion)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Generate a UUID for this database.
	id := uuid.New()
	_, err = tx.Exec(`INSERT INTO settings VALUES ('uuid', ?)`, id.String())
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	return err
}
