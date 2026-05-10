package database

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func Connect(dsn string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	return db, nil
}

// Migrate runs all SQL files in the given directory in lexicographic order.
// Each file is idempotent by convention (IF NOT EXISTS guards).
func Migrate(db *sqlx.DB, dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	sort.Strings(files)

	for _, f := range files {
		sql, err := os.ReadFile(f)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}
		if _, err := db.Exec(string(sql)); err != nil {
			return fmt.Errorf("run migration %s: %w", f, err)
		}
	}
	return nil
}
