package configs

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func Connect(config *Config) (*sql.DB, error) {
	db, err := sql.Open(config.Database.Driver, config.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %w", err)
	}

	return db, nil
}

func Migrate(db *sql.DB) error {
	// Read schema file.
	// Assuming the application is run from the root directory.
	schema, err := os.ReadFile("database/schema.sql")
	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	// Execute schema.
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	if err := ensureCatalogIDColumn(db); err != nil {
		return err
	}

	if err := ensureOSTemplateKindColumn(db); err != nil {
		return err
	}

	if err := ensureSprintDateColumns(db); err != nil {
		return err
	}

	if err := ensureOSExtraColumns(db); err != nil {
		return err
	}

	if err := ensureTaskSortOrderColumn(db); err != nil {
		return err
	}

	if err := ensureTaskResponsavelColumn(db); err != nil {
		return err
	}

	if err := ensureTaskTemplateKindColumn(db); err != nil {
		return err
	}

	if err := ensureCatalogNewTechBackfill(db); err != nil {
		return err
	}

	return nil
}

func ensureTaskTemplateKindColumn(db *sql.DB) error {
	exists, err := columnExists(db, "tasks", "template_kind")
	if err != nil {
		return fmt.Errorf("failed to inspect tasks table: %w", err)
	}
	if exists {
		return nil
	}
	if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN template_kind TEXT NOT NULL DEFAULT 'os'`); err != nil {
		return fmt.Errorf("failed to add tasks.template_kind column: %w", err)
	}
	return nil
}

func ensureTaskResponsavelColumn(db *sql.DB) error {
	exists, err := columnExists(db, "tasks", "responsavel")
	if err != nil {
		return fmt.Errorf("failed to inspect tasks table: %w", err)
	}
	if exists {
		return nil
	}
	if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN responsavel TEXT NOT NULL DEFAULT ''`); err != nil {
		return fmt.Errorf("failed to add tasks.responsavel column: %w", err)
	}
	return nil
}

// ensureCatalogNewTechBackfill fills missing UST Nova Tecnologia as 1.5× UST Normal
// (standard uplift used by the official catalog for coded items).
func ensureCatalogNewTechBackfill(db *sql.DB) error {
	_, err := db.Exec(`
		UPDATE catalog
		SET ust_new_tech = CAST(ROUND(ust_normal * 1.5) AS INTEGER)
		WHERE ust_new_tech IS NULL AND ust_normal > 0
	`)
	if err != nil {
		return fmt.Errorf("failed to backfill catalog ust_new_tech: %w", err)
	}
	return nil
}

func ensureTaskSortOrderColumn(db *sql.DB) error {
	exists, err := columnExists(db, "tasks", "sort_order")
	if err != nil {
		return fmt.Errorf("failed to inspect tasks table: %w", err)
	}
	if exists {
		// Column already present: keep user-defined order; never re-backfill.
		return nil
	}

	if _, err := db.Exec(`ALTER TABLE tasks ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("failed to add tasks.sort_order column: %w", err)
	}

	rows, err := db.Query(`
		SELECT id, sprint_id FROM tasks
		WHERE deleted_at IS NULL AND sprint_id IS NOT NULL
		ORDER BY sprint_id ASC, created_at DESC, id DESC
	`)
	if err != nil {
		return fmt.Errorf("failed to query tasks for sort_order backfill: %w", err)
	}

	type row struct {
		id       int
		sprintID int64
	}
	var items []row
	for rows.Next() {
		var id int
		var sprintID sql.NullInt64
		if err := rows.Scan(&id, &sprintID); err != nil {
			rows.Close()
			return fmt.Errorf("failed to scan task for sort_order: %w", err)
		}
		if !sprintID.Valid {
			continue
		}
		items = append(items, row{id: id, sprintID: sprintID.Int64})
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	var currentSprint int64 = -1
	order := 0
	for _, item := range items {
		if item.sprintID != currentSprint {
			currentSprint = item.sprintID
			order = 0
		}
		if _, err := db.Exec(`UPDATE tasks SET sort_order = ? WHERE id = ?`, order, item.id); err != nil {
			return fmt.Errorf("failed to backfill task sort_order: %w", err)
		}
		order++
	}
	return nil
}

func ensureOSExtraColumns(db *sql.DB) error {
	for _, column := range []struct {
		name string
		ddl  string
	}{
		{"observation", "ALTER TABLE os ADD COLUMN observation TEXT NOT NULL DEFAULT ''"},
		{"sort_order", "ALTER TABLE os ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0"},
	} {
		exists, err := columnExists(db, "os", column.name)
		if err != nil {
			return fmt.Errorf("failed to inspect os table: %w", err)
		}
		if exists {
			continue
		}
		if _, err := db.Exec(column.ddl); err != nil {
			return fmt.Errorf("failed to add os.%s column: %w", column.name, err)
		}
	}

	// Backfill sort_order for existing rows (stable by id).
	if _, err := db.Exec(`UPDATE os SET sort_order = id WHERE deleted_at IS NULL AND (sort_order IS NULL OR sort_order = 0)`); err != nil {
		return fmt.Errorf("failed to backfill os.sort_order: %w", err)
	}
	return nil
}

func ensureSprintDateColumns(db *sql.DB) error {
	for _, column := range []string{"start_date", "end_date"} {
		exists, err := columnExists(db, "sprints", column)
		if err != nil {
			return fmt.Errorf("failed to inspect sprints table: %w", err)
		}
		if exists {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf("ALTER TABLE sprints ADD COLUMN %s DATE", column)); err != nil {
			return fmt.Errorf("failed to add sprints.%s column: %w", column, err)
		}
	}
	return nil
}

func ensureOSTemplateKindColumn(db *sql.DB) error {
	exists, err := columnExists(db, "os_templates", "kind")
	if err != nil {
		return fmt.Errorf("failed to inspect os_templates table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec("ALTER TABLE os_templates ADD COLUMN kind TEXT NOT NULL DEFAULT 'os'"); err != nil {
		return fmt.Errorf("failed to add os_templates.kind column: %w", err)
	}
	return nil
}

func ensureCatalogIDColumn(db *sql.DB) error {
	exists, err := columnExists(db, "catalog", "catalog_id")
	if err != nil {
		return fmt.Errorf("failed to inspect catalog table: %w", err)
	}
	if exists {
		return nil
	}

	if _, err := db.Exec("ALTER TABLE catalog ADD COLUMN catalog_id INTEGER REFERENCES catalogs(id)"); err != nil {
		return fmt.Errorf("failed to add catalog_id column: %w", err)
	}
	return nil
}

func columnExists(db *sql.DB, table, column string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return false, err
		}
		if name == column {
			return true, nil
		}
	}
	return false, rows.Err()
}

func Seed(db *sql.DB) error {
	// Seed only the shared template items (catalog_id IS NULL).
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM catalog WHERE catalog_id IS NULL").Scan(&count); err != nil {
		return fmt.Errorf("failed to check catalog table: %w", err)
	}

	if count > 0 {
		return nil // Already seeded
	}

	// Read seed file.
	seed, err := os.ReadFile("database/seed.sql")
	if err != nil {
		return fmt.Errorf("failed to read seed file: %w", err)
	}

	// Execute seed.
	if _, err := db.Exec(string(seed)); err != nil {
		return fmt.Errorf("failed to execute seed: %w", err)
	}

	return nil
}
