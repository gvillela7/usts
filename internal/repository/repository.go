package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type Project struct {
	ID   int
	Name string
}

type Catalog struct {
	ID          int
	Cod         string
	Description string
	UstNormal   int
	UstNewTech  int
}

type Sprint struct {
	ID        int
	Name      string
	ProjectID int
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}

type Task struct {
	ID        int
	Name      string
	ProjectID int
	SprintID  sql.NullInt64
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}

type OS struct {
	ID              int
	ProjectID       int
	CatalogID       int
	TaskID          int
	SprintID        sql.NullInt64
	USTItem         int
	USTRecurrence   int
	Complexity      int
	USTEnd          *int
	Kind            string
	Description     string
	TarefaCandidta  int
	CreatedAt       time.Time
	UpdatedAt       sql.NullTime
	DeletedAt       sql.NullTime
}

type OSWithCatalog struct {
	OS
	CatalogCod         string
	CatalogDescription string
}

type Repository interface {
	GetProjects() ([]Project, error)
	GetProject(id int) (*Project, error)
	CreateProject(name string) error
	UpdateProject(projectID int, name string) error
	GetCodCatalog() ([]Catalog, error)
	GetSprints(projectID int) ([]Sprint, error)
	GetAllSprints() ([]Sprint, error)
	GetSprint(id int) (*Sprint, error)
	CreateSprint(name string, projectID int) error
	UpdateSprint(sprintID int, name string) error
	GetTasks(projectID int) ([]Task, error)
	GetTasksBySprint(sprintID int) ([]Task, error)
	GetTask(id int) (*Task, error)
	CreateTask(name string, projectID int, sprintID sql.NullInt64) error
	CreateMultipleOS(items []OS) error
	GetOSBySprint(sprintID int) ([]OSWithCatalog, error)
	GetOSByTask(taskID int) ([]OSWithCatalog, error)
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) GetProjects() ([]Project, error) {
	rows, err := r.db.Query("SELECT id, name FROM projects")
	if err != nil {
		return nil, fmt.Errorf("failed to query projects: %w", err)
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, fmt.Errorf("failed to scan project: %w", err)
		}
		projects = append(projects, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return projects, nil
}

func (r *SQLiteRepository) CreateProject(name string) error {
	// Check if project name already exists
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM projects WHERE name = ?", name).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check project name: %w", err)
	}

	if count > 0 {
		return fmt.Errorf("projeto com nome '%s' já existe", name)
	}

	// Insert new project
	_, err = r.db.Exec("INSERT INTO projects (name) VALUES (?)", name)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) UpdateProject(projectID int, name string) error {
	_, err := r.db.Exec(
		"UPDATE projects SET name = ? WHERE id = ?",
		name,
		projectID,
	)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetCodCatalog() ([]Catalog, error) {
	rows, err := r.db.Query("SELECT id, cod, description, ust_normal, COALESCE(ust_new_tech, ust_normal) as ust_new_tech FROM catalog")
	if err != nil {
		return nil, fmt.Errorf("failed to query catalog: %w", err)
	}
	defer rows.Close()

	var catalogs []Catalog
	for rows.Next() {
		var c Catalog
		if err := rows.Scan(&c.ID, &c.Cod, &c.Description, &c.UstNormal, &c.UstNewTech); err != nil {
			return nil, fmt.Errorf("failed to scan catalog: %w", err)
		}
		catalogs = append(catalogs, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return catalogs, nil
}

func (r *SQLiteRepository) GetSprints(projectID int) ([]Sprint, error) {
	rows, err := r.db.Query("SELECT id, name, project_id FROM sprints WHERE project_id = ? AND deleted_at IS NULL", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sprints: %w", err)
	}
	defer rows.Close()

	var sprints []Sprint
	for rows.Next() {
		var s Sprint
		if err := rows.Scan(&s.ID, &s.Name, &s.ProjectID); err != nil {
			return nil, fmt.Errorf("failed to scan sprint: %w", err)
		}
		sprints = append(sprints, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return sprints, nil
}

func (r *SQLiteRepository) GetAllSprints() ([]Sprint, error) {
	rows, err := r.db.Query("SELECT id, name, project_id, created_at FROM sprints WHERE deleted_at IS NULL ORDER BY project_id, created_at DESC")
	if err != nil {
		return nil, fmt.Errorf("failed to query all sprints: %w", err)
	}
	defer rows.Close()

	var sprints []Sprint
	for rows.Next() {
		var s Sprint
		if err := rows.Scan(&s.ID, &s.Name, &s.ProjectID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan sprint: %w", err)
		}
		sprints = append(sprints, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return sprints, nil
}

func (r *SQLiteRepository) GetSprint(id int) (*Sprint, error) {
	var s Sprint
	err := r.db.QueryRow("SELECT id, name, project_id, created_at FROM sprints WHERE id = ? AND deleted_at IS NULL", id).Scan(&s.ID, &s.Name, &s.ProjectID, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sprint not found")
		}
		return nil, fmt.Errorf("failed to query sprint: %w", err)
	}
	return &s, nil
}

func (r *SQLiteRepository) CreateSprint(name string, projectID int) error {
	_, err := r.db.Exec(
		"INSERT INTO sprints (name, project_id, created_at) VALUES (?, ?, ?)",
		name,
		projectID,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create sprint: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) UpdateSprint(sprintID int, name string) error {
	_, err := r.db.Exec(
		"UPDATE sprints SET name = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		name,
		time.Now(),
		sprintID,
	)
	if err != nil {
		return fmt.Errorf("failed to update sprint: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetTasks(projectID int) ([]Task, error) {
	rows, err := r.db.Query("SELECT id, name, project_id, sprint_id FROM tasks WHERE project_id = ? AND deleted_at IS NULL", projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Name, &t.ProjectID, &t.SprintID); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

func (r *SQLiteRepository) GetTasksBySprint(sprintID int) ([]Task, error) {
	rows, err := r.db.Query("SELECT id, name, project_id, sprint_id FROM tasks WHERE sprint_id = ? AND deleted_at IS NULL", sprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by sprint: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Name, &t.ProjectID, &t.SprintID); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return tasks, nil
}

func (r *SQLiteRepository) GetTask(id int) (*Task, error) {
	var t Task
	err := r.db.QueryRow(
		"SELECT id, name, project_id, sprint_id, created_at, updated_at, deleted_at FROM tasks WHERE id = ? AND deleted_at IS NULL",
		id,
	).Scan(&t.ID, &t.Name, &t.ProjectID, &t.SprintID, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to query task: %w", err)
	}
	return &t, nil
}

func (r *SQLiteRepository) CreateTask(name string, projectID int, sprintID sql.NullInt64) error {
	_, err := r.db.Exec(
		"INSERT INTO tasks (name, project_id, sprint_id, created_at) VALUES (?, ?, ?, ?)",
		name,
		projectID,
		sprintID,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetProject(id int) (*Project, error) {
	var p Project
	err := r.db.QueryRow("SELECT id, name FROM projects WHERE id = ?", id).Scan(&p.ID, &p.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		return nil, fmt.Errorf("failed to query project: %w", err)
	}
	return &p, nil
}

func (r *SQLiteRepository) CreateMultipleOS(items []OS) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("INSERT INTO os (project_id, catalog_id, task_id, sprint_id, ust_item, ust_recurrence, complexity, ust_end, kind, description, tarefa_candidata, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		_, err := stmt.Exec(
			item.ProjectID,
			item.CatalogID,
			item.TaskID,
			item.SprintID,
			item.USTItem,
			item.USTRecurrence,
			item.Complexity,
			item.USTEnd,
			item.Kind,
			item.Description,
			item.TarefaCandidta,
			time.Now(),
		)
		if err != nil {
			return fmt.Errorf("failed to insert OS item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *SQLiteRepository) GetOSBySprint(sprintID int) ([]OSWithCatalog, error) {
	rows, err := r.db.Query(`
		SELECT
			os.id, os.project_id, os.catalog_id, os.task_id, os.sprint_id,
			os.ust_item, os.ust_recurrence, os.complexity, os.ust_end,
			os.kind, os.description, os.tarefa_candidata, os.created_at,
			catalog.cod, catalog.description
		FROM os
		INNER JOIN catalog ON os.catalog_id = catalog.id
		WHERE os.sprint_id = ? AND os.deleted_at IS NULL
		ORDER BY os.created_at DESC
	`, sprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OS by sprint: %w", err)
	}
	defer rows.Close()

	var osList []OSWithCatalog
	for rows.Next() {
		var osWithCatalog OSWithCatalog
		if err := rows.Scan(
			&osWithCatalog.OS.ID,
			&osWithCatalog.OS.ProjectID,
			&osWithCatalog.OS.CatalogID,
			&osWithCatalog.OS.TaskID,
			&osWithCatalog.OS.SprintID,
			&osWithCatalog.OS.USTItem,
			&osWithCatalog.OS.USTRecurrence,
			&osWithCatalog.OS.Complexity,
			&osWithCatalog.OS.USTEnd,
			&osWithCatalog.OS.Kind,
			&osWithCatalog.OS.Description,
			&osWithCatalog.OS.TarefaCandidta,
			&osWithCatalog.OS.CreatedAt,
			&osWithCatalog.CatalogCod,
			&osWithCatalog.CatalogDescription,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OS: %w", err)
		}
		osList = append(osList, osWithCatalog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return osList, nil
}

func (r *SQLiteRepository) GetOSByTask(taskID int) ([]OSWithCatalog, error) {
	rows, err := r.db.Query(`
		SELECT
			os.id, os.project_id, os.catalog_id, os.task_id, os.sprint_id,
			os.ust_item, os.ust_recurrence, os.complexity, os.ust_end,
			os.kind, os.description, os.tarefa_candidata, os.created_at,
			catalog.cod, catalog.description
		FROM os
		INNER JOIN catalog ON os.catalog_id = catalog.id
		WHERE os.task_id = ? AND os.deleted_at IS NULL
		ORDER BY os.created_at DESC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OS by task: %w", err)
	}
	defer rows.Close()

	var osList []OSWithCatalog
	for rows.Next() {
		var osWithCatalog OSWithCatalog
		if err := rows.Scan(
			&osWithCatalog.OS.ID,
			&osWithCatalog.OS.ProjectID,
			&osWithCatalog.OS.CatalogID,
			&osWithCatalog.OS.TaskID,
			&osWithCatalog.OS.SprintID,
			&osWithCatalog.OS.USTItem,
			&osWithCatalog.OS.USTRecurrence,
			&osWithCatalog.OS.Complexity,
			&osWithCatalog.OS.USTEnd,
			&osWithCatalog.OS.Kind,
			&osWithCatalog.OS.Description,
			&osWithCatalog.OS.TarefaCandidta,
			&osWithCatalog.OS.CreatedAt,
			&osWithCatalog.CatalogCod,
			&osWithCatalog.CatalogDescription,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OS: %w", err)
		}
		osList = append(osList, osWithCatalog)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return osList, nil
}
