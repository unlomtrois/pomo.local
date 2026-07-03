package store

import (
	"errors"

	sqlite "modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// isUniqueViolation reports whether err is a SQLite UNIQUE/PRIMARY KEY
// constraint failure. The modernc driver surfaces these as *sqlite.Error with
// an extended result code, so we match on the constraint family rather than
// scraping the error string.
func isUniqueViolation(err error) bool {
	var se *sqlite.Error
	if !errors.As(err, &se) {
		return false
	}
	// Primary code is the low 8 bits of the extended code.
	return se.Code()&0xff == sqlite3.SQLITE_CONSTRAINT
}
