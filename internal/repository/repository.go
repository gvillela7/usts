package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type Project struct {
	ID   int
	Name string
}

type Catalog struct {
	ID          int
	CatalogID   int
	Cod         string
	Description string
	UstNormal   int
	UstNewTech  int
}

type ServiceCatalog struct {
	ID          int
	ProjectID   int
	ProjectName string
	Name        string
	StartDate   time.Time
	EndDate     time.Time
	ItemCount   int
	CreatedAt   time.Time
}

func (c ServiceCatalog) IsActive(now time.Time) bool {
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	start := time.Date(c.StartDate.Year(), c.StartDate.Month(), c.StartDate.Day(), 0, 0, 0, 0, today.Location())
	end := time.Date(c.EndDate.Year(), c.EndDate.Month(), c.EndDate.Day(), 0, 0, 0, 0, today.Location())
	return !today.Before(start) && !today.After(end)
}

type Sprint struct {
	ID        int
	Name      string
	ProjectID int
	StartDate time.Time
	EndDate   time.Time
	CreatedAt time.Time
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}

func (s Sprint) ListTitle() string {
	if s.StartDate.IsZero() || s.EndDate.IsZero() {
		return s.Name
	}
	return fmt.Sprintf("%s - %s - %s", s.Name, s.StartDate.Format("02/01"), s.EndDate.Format("02/01"))
}

func (s Sprint) Status() string {
	if s.StartDate.IsZero() || s.EndDate.IsZero() {
		return ""
	}
	today := time.Now()
	today = time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	start := time.Date(s.StartDate.Year(), s.StartDate.Month(), s.StartDate.Day(), 0, 0, 0, 0, today.Location())
	end := time.Date(s.EndDate.Year(), s.EndDate.Month(), s.EndDate.Day(), 0, 0, 0, 0, today.Location())
	if today.Before(start) {
		return "Planejada"
	}
	if today.After(end) {
		return "Encerrada"
	}
	return "Em andamento"
}

func (s Sprint) StartDateInput() string {
	if s.StartDate.IsZero() {
		return ""
	}
	return s.StartDate.Format("2006-01-02")
}

func (s Sprint) EndDateInput() string {
	if s.EndDate.IsZero() {
		return ""
	}
	return s.EndDate.Format("2006-01-02")
}

type Task struct {
	ID           int
	Name         string
	ProjectID    int
	SprintID     sql.NullInt64
	SortOrder    int
	Responsavel  string
	TemplateKind string
	CreatedAt    time.Time
	UpdatedAt    sql.NullTime
	DeletedAt    sql.NullTime
}

func (t Task) CreatedAtLabel() string {
	if t.CreatedAt.IsZero() {
		return "—"
	}
	return t.CreatedAt.Format("02/01/2006")
}

func (t Task) CreatedAtInput() string {
	if t.CreatedAt.IsZero() {
		return time.Now().Format("2006-01-02")
	}
	return t.CreatedAt.Format("2006-01-02")
}

type OS struct {
	ID             int
	ProjectID      int
	CatalogID      int
	TaskID         int
	SprintID       sql.NullInt64
	USTItem        int
	USTRecurrence  int
	Complexity     int
	USTEnd         *int
	Kind           string
	Description    string
	Observation    string
	SortOrder      int
	TarefaCandidta int
	CreatedAt      time.Time
	UpdatedAt      sql.NullTime
	DeletedAt      sql.NullTime
}

type OSWithCatalog struct {
	OS
	CatalogCod         string
	CatalogDescription string
}

const (
	OSTemplateKindOS    = "os"
	OSTemplateKindEpico = "epico"
)

type OSTemplate struct {
	ID        int
	ProjectID int
	Name      string
	Kind      string
	ItemCount int
	CreatedAt time.Time
}

func NormalizeOSTemplateKind(kind string) string {
	if kind == OSTemplateKindEpico {
		return OSTemplateKindEpico
	}
	return OSTemplateKindOS
}

func (t OSTemplate) KindLabel() string {
	if t.Kind == OSTemplateKindEpico {
		return "Template OS Épico"
	}
	return "Template OS"
}

type OSTemplateItem struct {
	ID             int
	TemplateID     int
	CatalogID      int
	CatalogCod     string
	CatalogDesc    string
	USTItem        int
	USTRecurrence  int
	Complexity     int
	USTEnd         *int
	Kind           string
	Description    string
	TarefaCandidta int
	UstNormal      int
	UstNewTech     int
}

type Repository interface {
	GetProjects() ([]Project, error)
	GetProject(id int) (*Project, error)
	CreateProject(name string) error
	UpdateProject(projectID int, name string) error
	GetCodCatalog() ([]Catalog, error)
	GetServiceCatalogs() ([]ServiceCatalog, error)
	GetServiceCatalogsByProject(projectID int) ([]ServiceCatalog, error)
	GetServiceCatalog(id int) (*ServiceCatalog, error)
	CreateServiceCatalog(projectID int, name string, startDate, endDate time.Time) (int, error)
	HasTemplateCatalogItems() (bool, error)
	CopyTemplateItems(catalogID int) error
	GetCatalogItems(catalogID int) ([]Catalog, error)
	GetCatalogItemsByProject(projectID int) ([]Catalog, error)
	GetCatalogItem(id int) (*Catalog, error)
	CreateCatalogItem(catalogID int, cod, description string, ustNormal int, ustNewTech *int, unitMeasurement string) error
	UpdateCatalogItem(id, catalogID int, cod, description string, ustNormal int, ustNewTech *int) error
	GetSprints(projectID int) ([]Sprint, error)
	GetAllSprints() ([]Sprint, error)
	GetSprint(id int) (*Sprint, error)
	CreateSprint(name string, projectID int, startDate, endDate time.Time) error
	UpdateSprint(sprintID int, name string, startDate, endDate time.Time) error
	GetTasks(projectID int) ([]Task, error)
	GetTasksBySprint(sprintID int) ([]Task, error)
	GetTask(id int) (*Task, error)
	CreateTask(name string, projectID int, sprintID sql.NullInt64) error
	UpdateTask(id int, name string) error
	UpdateTaskResponsavel(id int, responsavel string) error
	UpdateTaskTemplateKind(id int, templateKind string) error
	ReorderSprintTasks(sprintID int, taskIDs []int) error
	FindOrCreateTask(name string, projectID, sprintID int) (int, error)
	CreateMultipleOS(items []OS) error
	SaveTaskOSItems(taskID int, items []OS) error
	GetOSBySprint(sprintID int) ([]OSWithCatalog, error)
	GetOSByTask(taskID int) ([]OSWithCatalog, error)
	GetOSTemplates(projectID int) ([]OSTemplate, error)
	GetOSTemplate(id int) (*OSTemplate, error)
	CreateOSTemplate(projectID int, name, kind string, items []OSTemplateItem) (int, error)
	UpdateOSTemplate(id, projectID int, name, kind string, items []OSTemplateItem) error
	DeleteOSTemplate(id, projectID int) error
	GetOSTemplateItems(templateID int) ([]OSTemplateItem, error)
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
	return r.queryCatalogItems("SELECT id, COALESCE(catalog_id, 0), cod, description, ust_normal, COALESCE(ust_new_tech, ust_normal) as ust_new_tech FROM catalog WHERE catalog_id IS NULL ORDER BY cod")
}

func (r *SQLiteRepository) GetServiceCatalogs() ([]ServiceCatalog, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id, c.project_id, p.name, c.name, c.start_date, c.end_date, c.created_at,
			(SELECT COUNT(*) FROM catalog items WHERE items.catalog_id = c.id) AS item_count
		FROM catalogs c
		INNER JOIN projects p ON p.id = c.project_id
		WHERE c.deleted_at IS NULL
		ORDER BY p.name, c.start_date DESC, c.name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query catalogs: %w", err)
	}
	defer rows.Close()

	var catalogs []ServiceCatalog
	for rows.Next() {
		catalog, err := scanServiceCatalog(rows)
		if err != nil {
			return nil, err
		}
		catalogs = append(catalogs, catalog)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return catalogs, nil
}

func (r *SQLiteRepository) GetServiceCatalogsByProject(projectID int) ([]ServiceCatalog, error) {
	rows, err := r.db.Query(`
		SELECT
			c.id, c.project_id, p.name, c.name, c.start_date, c.end_date, c.created_at,
			(SELECT COUNT(*) FROM catalog items WHERE items.catalog_id = c.id) AS item_count
		FROM catalogs c
		INNER JOIN projects p ON p.id = c.project_id
		WHERE c.deleted_at IS NULL AND c.project_id = ?
		ORDER BY c.start_date DESC, c.name
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query project catalogs: %w", err)
	}
	defer rows.Close()

	var catalogs []ServiceCatalog
	for rows.Next() {
		catalog, err := scanServiceCatalog(rows)
		if err != nil {
			return nil, err
		}
		catalogs = append(catalogs, catalog)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return catalogs, nil
}

func (r *SQLiteRepository) GetServiceCatalog(id int) (*ServiceCatalog, error) {
	row := r.db.QueryRow(`
		SELECT
			c.id, c.project_id, p.name, c.name, c.start_date, c.end_date, c.created_at,
			(SELECT COUNT(*) FROM catalog items WHERE items.catalog_id = c.id) AS item_count
		FROM catalogs c
		INNER JOIN projects p ON p.id = c.project_id
		WHERE c.id = ? AND c.deleted_at IS NULL
	`, id)

	catalog, err := scanServiceCatalog(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("catalog not found")
		}
		return nil, fmt.Errorf("failed to scan catalog: %w", err)
	}
	return &catalog, nil
}

func (r *SQLiteRepository) CreateServiceCatalog(projectID int, name string, startDate, endDate time.Time) (int, error) {
	result, err := r.db.Exec(
		"INSERT INTO catalogs (project_id, name, start_date, end_date, created_at) VALUES (?, ?, ?, ?, ?)",
		projectID,
		name,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create catalog: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get catalog id: %w", err)
	}
	return int(id), nil
}

func (r *SQLiteRepository) HasTemplateCatalogItems() (bool, error) {
	var count int
	if err := r.db.QueryRow("SELECT COUNT(*) FROM catalog WHERE catalog_id IS NULL").Scan(&count); err != nil {
		return false, fmt.Errorf("failed to check template catalog items: %w", err)
	}
	return count > 0, nil
}

func (r *SQLiteRepository) CopyTemplateItems(catalogID int) error {
	_, err := r.db.Exec(`
		INSERT INTO catalog (catalog_id, att, cod, description, ust_normal, ust_adapt, ust_new_tech, unit_measurement)
		SELECT ?, att, cod, description, ust_normal, ust_adapt, ust_new_tech, unit_measurement
		FROM catalog
		WHERE catalog_id IS NULL
	`, catalogID)
	if err != nil {
		return fmt.Errorf("failed to copy template catalog items: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetCatalogItems(catalogID int) ([]Catalog, error) {
	return r.queryCatalogItems(`
		SELECT id, COALESCE(catalog_id, 0), cod, description, ust_normal, COALESCE(ust_new_tech, ust_normal) as ust_new_tech
		FROM catalog
		WHERE catalog_id = ?
		ORDER BY cod
	`, catalogID)
}

func (r *SQLiteRepository) GetCatalogItemsByProject(projectID int) ([]Catalog, error) {
	return r.queryCatalogItems(`
		SELECT items.id, COALESCE(items.catalog_id, 0), items.cod, items.description, items.ust_normal,
			COALESCE(items.ust_new_tech, items.ust_normal) as ust_new_tech
		FROM catalog items
		INNER JOIN catalogs c ON c.id = items.catalog_id
		WHERE c.project_id = ? AND c.deleted_at IS NULL
		ORDER BY items.cod
	`, projectID)
}

func (r *SQLiteRepository) CreateCatalogItem(catalogID int, cod, description string, ustNormal int, ustNewTech *int, unitMeasurement string) error {
	if unitMeasurement == "" {
		unitMeasurement = "Unidade"
	}
	_, err := r.db.Exec(
		"INSERT INTO catalog (catalog_id, cod, description, ust_normal, ust_new_tech, unit_measurement) VALUES (?, ?, ?, ?, ?, ?)",
		catalogID,
		cod,
		description,
		ustNormal,
		ustNewTech,
		unitMeasurement,
	)
	if err != nil {
		return fmt.Errorf("failed to create catalog item: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) GetCatalogItem(id int) (*Catalog, error) {
	var c Catalog
	err := r.db.QueryRow(`
		SELECT id, COALESCE(catalog_id, 0), cod, description, ust_normal, COALESCE(ust_new_tech, ust_normal)
		FROM catalog
		WHERE id = ?
	`, id).Scan(&c.ID, &c.CatalogID, &c.Cod, &c.Description, &c.UstNormal, &c.UstNewTech)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("catalog item not found")
		}
		return nil, fmt.Errorf("failed to query catalog item: %w", err)
	}
	return &c, nil
}

func (r *SQLiteRepository) UpdateCatalogItem(id, catalogID int, cod, description string, ustNormal int, ustNewTech *int) error {
	result, err := r.db.Exec(
		"UPDATE catalog SET cod = ?, description = ?, ust_normal = ?, ust_new_tech = ? WHERE id = ? AND catalog_id = ?",
		cod,
		description,
		ustNormal,
		ustNewTech,
		id,
		catalogID,
	)
	if err != nil {
		return fmt.Errorf("failed to update catalog item: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check catalog item update: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("catalog item not found")
	}
	return nil
}

type catalogScanner interface {
	Scan(dest ...any) error
}

func scanServiceCatalog(scanner catalogScanner) (ServiceCatalog, error) {
	var catalog ServiceCatalog
	var startDate, endDate string
	if err := scanner.Scan(
		&catalog.ID,
		&catalog.ProjectID,
		&catalog.ProjectName,
		&catalog.Name,
		&startDate,
		&endDate,
		&catalog.CreatedAt,
		&catalog.ItemCount,
	); err != nil {
		return ServiceCatalog{}, err
	}

	parsedStart, err := parseFlexibleDate(startDate)
	if err != nil {
		return ServiceCatalog{}, fmt.Errorf("failed to parse start date: %w", err)
	}
	parsedEnd, err := parseFlexibleDate(endDate)
	if err != nil {
		return ServiceCatalog{}, fmt.Errorf("failed to parse end date: %w", err)
	}
	catalog.StartDate = parsedStart
	catalog.EndDate = parsedEnd
	return catalog, nil
}

func parseFlexibleDate(value string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
	}
	var lastErr error
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return parsed, nil
		}
		lastErr = err
	}
	return time.Time{}, lastErr
}

func (r *SQLiteRepository) queryCatalogItems(query string, args ...any) ([]Catalog, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query catalog: %w", err)
	}
	defer rows.Close()

	var catalogs []Catalog
	for rows.Next() {
		var c Catalog
		if err := rows.Scan(&c.ID, &c.CatalogID, &c.Cod, &c.Description, &c.UstNormal, &c.UstNewTech); err != nil {
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
	rows, err := r.db.Query(`
		SELECT id, name, project_id, start_date, end_date, created_at
		FROM sprints
		WHERE project_id = ? AND deleted_at IS NULL
		ORDER BY COALESCE(start_date, created_at) DESC, name
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sprints: %w", err)
	}
	defer rows.Close()

	var sprints []Sprint
	for rows.Next() {
		s, err := scanSprint(rows)
		if err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return sprints, nil
}

func (r *SQLiteRepository) GetAllSprints() ([]Sprint, error) {
	rows, err := r.db.Query(`
		SELECT id, name, project_id, start_date, end_date, created_at
		FROM sprints
		WHERE deleted_at IS NULL
		ORDER BY project_id, COALESCE(start_date, created_at) DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all sprints: %w", err)
	}
	defer rows.Close()

	var sprints []Sprint
	for rows.Next() {
		s, err := scanSprint(rows)
		if err != nil {
			return nil, err
		}
		sprints = append(sprints, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return sprints, nil
}

func (r *SQLiteRepository) GetSprint(id int) (*Sprint, error) {
	s, err := scanSprintRow(r.db.QueryRow(`
		SELECT id, name, project_id, start_date, end_date, created_at
		FROM sprints
		WHERE id = ? AND deleted_at IS NULL
	`, id))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sprint not found")
		}
		return nil, fmt.Errorf("failed to query sprint: %w", err)
	}
	return &s, nil
}

func (r *SQLiteRepository) CreateSprint(name string, projectID int, startDate, endDate time.Time) error {
	_, err := r.db.Exec(
		"INSERT INTO sprints (name, project_id, start_date, end_date, created_at) VALUES (?, ?, ?, ?, ?)",
		name,
		projectID,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create sprint: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) UpdateSprint(sprintID int, name string, startDate, endDate time.Time) error {
	_, err := r.db.Exec(
		"UPDATE sprints SET name = ?, start_date = ?, end_date = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		name,
		startDate.Format("2006-01-02"),
		endDate.Format("2006-01-02"),
		time.Now(),
		sprintID,
	)
	if err != nil {
		return fmt.Errorf("failed to update sprint: %w", err)
	}
	return nil
}

type sprintScanner interface {
	Scan(dest ...any) error
}

func scanSprint(rows *sql.Rows) (Sprint, error) {
	return scanSprintRow(rows)
}

func scanSprintRow(row sprintScanner) (Sprint, error) {
	var s Sprint
	var startStr, endStr sql.NullString
	if err := row.Scan(&s.ID, &s.Name, &s.ProjectID, &startStr, &endStr, &s.CreatedAt); err != nil {
		return Sprint{}, err
	}
	if startStr.Valid {
		if parsed, err := parseFlexibleDate(startStr.String); err == nil {
			s.StartDate = parsed
		}
	}
	if endStr.Valid {
		if parsed, err := parseFlexibleDate(endStr.String); err == nil {
			s.EndDate = parsed
		}
	}
	return s, nil
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
	rows, err := r.db.Query(
		`SELECT id, name, project_id, sprint_id, COALESCE(sort_order, 0), COALESCE(responsavel, ''), COALESCE(template_kind, 'os'), created_at
		 FROM tasks
		 WHERE sprint_id = ? AND deleted_at IS NULL
		 ORDER BY COALESCE(sort_order, 0) ASC, id ASC`,
		sprintID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks by sprint: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Name, &t.ProjectID, &t.SprintID, &t.SortOrder, &t.Responsavel, &t.TemplateKind, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}
		t.TemplateKind = NormalizeOSTemplateKind(t.TemplateKind)
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
		`SELECT id, name, project_id, sprint_id, COALESCE(responsavel, ''), COALESCE(template_kind, 'os'), created_at, updated_at, deleted_at
		 FROM tasks WHERE id = ? AND deleted_at IS NULL`,
		id,
	).Scan(&t.ID, &t.Name, &t.ProjectID, &t.SprintID, &t.Responsavel, &t.TemplateKind, &t.CreatedAt, &t.UpdatedAt, &t.DeletedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to query task: %w", err)
	}
	t.TemplateKind = NormalizeOSTemplateKind(t.TemplateKind)
	return &t, nil
}

func (r *SQLiteRepository) CreateTask(name string, projectID int, sprintID sql.NullInt64) error {
	sortOrder := 0
	if sprintID.Valid {
		_ = r.db.QueryRow(
			`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM tasks WHERE sprint_id = ? AND deleted_at IS NULL`,
			sprintID.Int64,
		).Scan(&sortOrder)
	}
	_, err := r.db.Exec(
		"INSERT INTO tasks (name, project_id, sprint_id, sort_order, responsavel, template_kind, created_at) VALUES (?, ?, ?, ?, '', 'os', ?)",
		name,
		projectID,
		sprintID,
		sortOrder,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to create task: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) UpdateTask(id int, name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return fmt.Errorf("nome da tarefa é obrigatório")
	}
	res, err := r.db.Exec(
		"UPDATE tasks SET name = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		name,
		time.Now(),
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check task update: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("tarefa não encontrada")
	}
	return nil
}

func (r *SQLiteRepository) UpdateTaskResponsavel(id int, responsavel string) error {
	responsavel = strings.TrimSpace(responsavel)
	if len([]rune(responsavel)) > 120 {
		runes := []rune(responsavel)
		responsavel = string(runes[:120])
	}
	res, err := r.db.Exec(
		"UPDATE tasks SET responsavel = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		responsavel,
		time.Now(),
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task responsavel: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check task responsavel update: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("tarefa não encontrada")
	}
	return nil
}

func (r *SQLiteRepository) UpdateTaskTemplateKind(id int, templateKind string) error {
	templateKind = NormalizeOSTemplateKind(templateKind)
	res, err := r.db.Exec(
		"UPDATE tasks SET template_kind = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",
		templateKind,
		time.Now(),
		id,
	)
	if err != nil {
		return fmt.Errorf("failed to update task template_kind: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check task template_kind update: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("tarefa não encontrada")
	}
	return nil
}

func (r *SQLiteRepository) ReorderSprintTasks(sprintID int, taskIDs []int) error {
	if sprintID <= 0 {
		return fmt.Errorf("sprint inválida")
	}
	if len(taskIDs) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin task reorder: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for i, taskID := range taskIDs {
		res, err := tx.Exec(
			`UPDATE tasks SET sort_order = ?, updated_at = ?
			 WHERE id = ? AND sprint_id = ? AND deleted_at IS NULL`,
			i,
			time.Now(),
			taskID,
			sprintID,
		)
		if err != nil {
			return fmt.Errorf("failed to update task order: %w", err)
		}
		n, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check task order update: %w", err)
		}
		if n == 0 {
			return fmt.Errorf("tarefa %d não pertence à sprint", taskID)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit task reorder: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) FindOrCreateTask(name string, projectID, sprintID int) (int, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return 0, fmt.Errorf("nome da tarefa é obrigatório")
	}

	var id int
	err := r.db.QueryRow(
		`SELECT id FROM tasks
		 WHERE name = ? AND project_id = ? AND sprint_id = ? AND deleted_at IS NULL
		 ORDER BY id DESC LIMIT 1`,
		name, projectID, sprintID,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to query task: %w", err)
	}

	var nextSort int
	_ = r.db.QueryRow(
		`SELECT COALESCE(MAX(sort_order), -1) + 1 FROM tasks WHERE sprint_id = ? AND deleted_at IS NULL`,
		sprintID,
	).Scan(&nextSort)

	res, err := r.db.Exec(
		"INSERT INTO tasks (name, project_id, sprint_id, sort_order, responsavel, template_kind, created_at) VALUES (?, ?, ?, ?, '', 'os', ?)",
		name,
		projectID,
		sprintID,
		nextSort,
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create task: %w", err)
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get task id: %w", err)
	}
	return int(lastID), nil
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

	stmt, err := tx.Prepare(`
		INSERT INTO os (
			project_id, catalog_id, task_id, sprint_id,
			ust_item, ust_recurrence, complexity, ust_end,
			kind, description, observation, sort_order, tarefa_candidata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, item := range items {
		sortOrder := item.SortOrder
		if sortOrder <= 0 {
			sortOrder = i + 1
		}
		observation := item.Observation
		if len(observation) > 1000 {
			observation = observation[:1000]
		}
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
			observation,
			sortOrder,
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

func (r *SQLiteRepository) SaveTaskOSItems(taskID int, items []OS) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	keepIDs := make([]int, 0, len(items))
	for _, item := range items {
		if item.ID > 0 {
			keepIDs = append(keepIDs, item.ID)
		}
	}

	if len(keepIDs) == 0 {
		if _, err := tx.Exec(
			"UPDATE os SET deleted_at = ?, updated_at = ? WHERE task_id = ? AND deleted_at IS NULL",
			time.Now(), time.Now(), taskID,
		); err != nil {
			return fmt.Errorf("failed to soft-delete OS items: %w", err)
		}
	} else {
		placeholders := make([]string, len(keepIDs))
		args := make([]any, 0, len(keepIDs)+3)
		args = append(args, time.Now(), time.Now(), taskID)
		for i, id := range keepIDs {
			placeholders[i] = "?"
			args = append(args, id)
		}
		query := fmt.Sprintf(
			`UPDATE os SET deleted_at = ?, updated_at = ?
			 WHERE task_id = ? AND deleted_at IS NULL AND id NOT IN (%s)`,
			strings.Join(placeholders, ","),
		)
		if _, err := tx.Exec(query, args...); err != nil {
			return fmt.Errorf("failed to soft-delete removed OS items: %w", err)
		}
	}

	updateStmt, err := tx.Prepare(`
		UPDATE os SET
			catalog_id = ?, ust_item = ?, ust_recurrence = ?, complexity = ?, ust_end = ?,
			kind = ?, description = ?, observation = ?, sort_order = ?, tarefa_candidata = ?, updated_at = ?
		WHERE id = ? AND task_id = ? AND deleted_at IS NULL
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare OS update: %w", err)
	}
	defer updateStmt.Close()

	insertStmt, err := tx.Prepare(`
		INSERT INTO os (
			project_id, catalog_id, task_id, sprint_id,
			ust_item, ust_recurrence, complexity, ust_end,
			kind, description, observation, sort_order, tarefa_candidata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare OS insert: %w", err)
	}
	defer insertStmt.Close()

	now := time.Now()
	for i, item := range items {
		observation := strings.TrimSpace(item.Observation)
		if len([]rune(observation)) > 1000 {
			runes := []rune(observation)
			observation = string(runes[:1000])
		}
		sortOrder := i + 1
		kind := NormalizeOSKind(item.Kind)

		if item.ID > 0 {
			_, err := updateStmt.Exec(
				item.CatalogID,
				item.USTItem,
				item.USTRecurrence,
				item.Complexity,
				item.USTEnd,
				kind,
				item.Description,
				observation,
				sortOrder,
				item.TarefaCandidta,
				now,
				item.ID,
				taskID,
			)
			if err != nil {
				return fmt.Errorf("failed to update OS item %d: %w", item.ID, err)
			}
			continue
		}

		_, err := insertStmt.Exec(
			item.ProjectID,
			item.CatalogID,
			taskID,
			item.SprintID,
			item.USTItem,
			item.USTRecurrence,
			item.Complexity,
			item.USTEnd,
			kind,
			item.Description,
			observation,
			sortOrder,
			item.TarefaCandidta,
			now,
		)
		if err != nil {
			return fmt.Errorf("failed to insert OS item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit OS save: %w", err)
	}
	return nil
}

func NormalizeOSKind(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "nova_tecnologia", "nova tecnologia", "ajuste":
		return "nova_tecnologia"
	default:
		return "normal"
	}
}

func (r *SQLiteRepository) GetOSBySprint(sprintID int) ([]OSWithCatalog, error) {
	rows, err := r.db.Query(`
		SELECT
			os.id, os.project_id, os.catalog_id, os.task_id, os.sprint_id,
			os.ust_item, os.ust_recurrence, os.complexity, os.ust_end,
			os.kind, os.description, COALESCE(os.observation, ''), COALESCE(os.sort_order, 0),
			os.tarefa_candidata, os.created_at,
			catalog.cod, catalog.description
		FROM os
		INNER JOIN catalog ON os.catalog_id = catalog.id
		WHERE os.sprint_id = ? AND os.deleted_at IS NULL
		ORDER BY COALESCE(os.sort_order, 0) ASC, os.id ASC
	`, sprintID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OS by sprint: %w", err)
	}
	defer rows.Close()

	return scanOSWithCatalogRows(rows)
}

func (r *SQLiteRepository) GetOSByTask(taskID int) ([]OSWithCatalog, error) {
	rows, err := r.db.Query(`
		SELECT
			os.id, os.project_id, os.catalog_id, os.task_id, os.sprint_id,
			os.ust_item, os.ust_recurrence, os.complexity, os.ust_end,
			os.kind, os.description, COALESCE(os.observation, ''), COALESCE(os.sort_order, 0),
			os.tarefa_candidata, os.created_at,
			catalog.cod, catalog.description
		FROM os
		INNER JOIN catalog ON os.catalog_id = catalog.id
		WHERE os.task_id = ? AND os.deleted_at IS NULL
		ORDER BY COALESCE(os.sort_order, 0) ASC, os.id ASC
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OS by task: %w", err)
	}
	defer rows.Close()

	return scanOSWithCatalogRows(rows)
}

func scanOSWithCatalogRows(rows *sql.Rows) ([]OSWithCatalog, error) {
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
			&osWithCatalog.OS.Observation,
			&osWithCatalog.OS.SortOrder,
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

func (r *SQLiteRepository) GetOSTemplates(projectID int) ([]OSTemplate, error) {
	rows, err := r.db.Query(`
		SELECT t.id, t.project_id, t.name, COALESCE(t.kind, 'os'), t.created_at,
			(SELECT COUNT(*) FROM os_template_items i WHERE i.template_id = t.id) as item_count
		FROM os_templates t
		WHERE t.project_id = ? AND t.deleted_at IS NULL
		ORDER BY CASE t.kind WHEN 'epico' THEN 0 ELSE 1 END, t.name
	`, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OS templates: %w", err)
	}
	defer rows.Close()

	var templates []OSTemplate
	for rows.Next() {
		var t OSTemplate
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Name, &t.Kind, &t.CreatedAt, &t.ItemCount); err != nil {
			return nil, fmt.Errorf("failed to scan OS template: %w", err)
		}
		t.Kind = NormalizeOSTemplateKind(t.Kind)
		templates = append(templates, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return templates, nil
}

func (r *SQLiteRepository) GetOSTemplate(id int) (*OSTemplate, error) {
	var t OSTemplate
	err := r.db.QueryRow(`
		SELECT t.id, t.project_id, t.name, COALESCE(t.kind, 'os'), t.created_at,
			(SELECT COUNT(*) FROM os_template_items i WHERE i.template_id = t.id) as item_count
		FROM os_templates t
		WHERE t.id = ? AND t.deleted_at IS NULL
	`, id).Scan(&t.ID, &t.ProjectID, &t.Name, &t.Kind, &t.CreatedAt, &t.ItemCount)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("OS template not found")
		}
		return nil, fmt.Errorf("failed to query OS template: %w", err)
	}
	t.Kind = NormalizeOSTemplateKind(t.Kind)
	return &t, nil
}

func (r *SQLiteRepository) GetOSTemplateItems(templateID int) ([]OSTemplateItem, error) {
	rows, err := r.db.Query(`
		SELECT
			i.id, i.template_id, i.catalog_id,
			c.cod, c.description,
			i.ust_item, i.ust_recurrence, i.complexity, i.ust_end,
			i.kind, i.description, i.tarefa_candidata,
			c.ust_normal, COALESCE(c.ust_new_tech, c.ust_normal)
		FROM os_template_items i
		INNER JOIN catalog c ON c.id = i.catalog_id
		WHERE i.template_id = ?
		ORDER BY i.id
	`, templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to query OS template items: %w", err)
	}
	defer rows.Close()

	var items []OSTemplateItem
	for rows.Next() {
		var item OSTemplateItem
		var ustEnd sql.NullInt64
		if err := rows.Scan(
			&item.ID,
			&item.TemplateID,
			&item.CatalogID,
			&item.CatalogCod,
			&item.CatalogDesc,
			&item.USTItem,
			&item.USTRecurrence,
			&item.Complexity,
			&ustEnd,
			&item.Kind,
			&item.Description,
			&item.TarefaCandidta,
			&item.UstNormal,
			&item.UstNewTech,
		); err != nil {
			return nil, fmt.Errorf("failed to scan OS template item: %w", err)
		}
		if ustEnd.Valid {
			v := int(ustEnd.Int64)
			item.USTEnd = &v
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}
	return items, nil
}

func (r *SQLiteRepository) CreateOSTemplate(projectID int, name, kind string, items []OSTemplateItem) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"INSERT INTO os_templates (project_id, name, kind, created_at) VALUES (?, ?, ?, ?)",
		projectID,
		name,
		NormalizeOSTemplateKind(kind),
		time.Now(),
	)
	if err != nil {
		return 0, fmt.Errorf("failed to create OS template: %w", err)
	}

	templateID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get OS template id: %w", err)
	}

	if err := insertOSTemplateItems(tx, int(templateID), items); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit OS template: %w", err)
	}
	return int(templateID), nil
}

func (r *SQLiteRepository) UpdateOSTemplate(id, projectID int, name, kind string, items []OSTemplateItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"UPDATE os_templates SET name = ?, kind = ?, updated_at = ? WHERE id = ? AND project_id = ? AND deleted_at IS NULL",
		name,
		NormalizeOSTemplateKind(kind),
		time.Now(),
		id,
		projectID,
	)
	if err != nil {
		return fmt.Errorf("failed to update OS template: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check OS template update: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("OS template not found")
	}

	if _, err := tx.Exec("DELETE FROM os_template_items WHERE template_id = ?", id); err != nil {
		return fmt.Errorf("failed to clear OS template items: %w", err)
	}

	if err := insertOSTemplateItems(tx, id, items); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit OS template update: %w", err)
	}
	return nil
}

func (r *SQLiteRepository) DeleteOSTemplate(id, projectID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		"UPDATE os_templates SET deleted_at = ? WHERE id = ? AND project_id = ? AND deleted_at IS NULL",
		time.Now(),
		id,
		projectID,
	)
	if err != nil {
		return fmt.Errorf("failed to delete OS template: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check OS template delete: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf("OS template not found")
	}

	if _, err := tx.Exec("DELETE FROM os_template_items WHERE template_id = ?", id); err != nil {
		return fmt.Errorf("failed to delete OS template items: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit OS template delete: %w", err)
	}
	return nil
}

func insertOSTemplateItems(tx *sql.Tx, templateID int, items []OSTemplateItem) error {
	stmt, err := tx.Prepare(`
		INSERT INTO os_template_items (
			template_id, catalog_id, ust_item, ust_recurrence, complexity,
			ust_end, kind, description, tarefa_candidata, created_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare OS template item insert: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	for _, item := range items {
		if item.CatalogID <= 0 {
			return fmt.Errorf("template item requires catalog_id")
		}
		if _, err := stmt.Exec(
			templateID,
			item.CatalogID,
			1,
			1,
			1,
			nil,
			"normal",
			"",
			0,
			now,
		); err != nil {
			return fmt.Errorf("failed to insert OS template item: %w", err)
		}
	}
	return nil
}
