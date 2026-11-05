CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS catalogs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME,
    FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS catalog (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    catalog_id INTEGER,
    att TEXT,
    cod TEXT NOT NULL,
    description TEXT NOT NULL,
    ust_normal INTEGER NOT NULL,
    ust_adapt INTEGER,
    ust_new_tech INTEGER,
    unit_measurement TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(catalog_id) REFERENCES catalogs(id)
);

CREATE TABLE IF NOT EXISTS sprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    project_id INT NOT NULL,
    start_date DATE,
    end_date DATE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME,
    FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    project_id INT NOT NULL,
    sprint_id INTEGER,
    sort_order INTEGER NOT NULL DEFAULT 0,
    responsavel TEXT NOT NULL DEFAULT '',
    template_kind TEXT NOT NULL DEFAULT 'os',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME,
    FOREIGN KEY(project_id) REFERENCES projects(id),
    FOREIGN KEY(sprint_id) REFERENCES sprints(id)
);

CREATE TABLE IF NOT EXISTS os (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER,
    catalog_id INTEGER,
    task_id INTEGER,
    sprint_id INTEGER,
    ust_item INTEGER NOT NULL,
    ust_recurrence INTEGER NOT NULL,
    complexity INTEGER NOT NULL,
    ust_end INTEGER,
    kind TEXT NOT NULL,
    description TEXT NOT NULL,
    observation TEXT NOT NULL DEFAULT '',
    sort_order INTEGER NOT NULL DEFAULT 0,
    tarefa_candidata INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME,
    FOREIGN KEY(project_id) REFERENCES projects(id),
    FOREIGN KEY(catalog_id) REFERENCES catalog(id),
    FOREIGN KEY(task_id) REFERENCES tasks(id),
    FOREIGN KEY(sprint_id) REFERENCES sprints(id)
);

CREATE TABLE IF NOT EXISTS os_templates (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    kind TEXT NOT NULL DEFAULT 'os',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME,
    FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS os_template_items (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    template_id INTEGER NOT NULL,
    catalog_id INTEGER NOT NULL,
    ust_item INTEGER NOT NULL,
    ust_recurrence INTEGER NOT NULL DEFAULT 1,
    complexity INTEGER NOT NULL DEFAULT 1,
    ust_end INTEGER,
    kind TEXT NOT NULL DEFAULT 'normal',
    description TEXT NOT NULL,
    tarefa_candidata INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY(template_id) REFERENCES os_templates(id),
    FOREIGN KEY(catalog_id) REFERENCES catalog(id)
);
