package database

import (
	"database/sql"
	"encoding/json"
)

// NullString wraps sql.NullString so it marshals to a JSON string or null
// rather than the default {"String":"","Valid":false} struct encoding.
type NullString struct {
	sql.NullString
}

func (ns NullString) MarshalJSON() ([]byte, error) {
	if !ns.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(ns.String)
}
