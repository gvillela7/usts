CREATE TABLE IF NOT EXISTS catalog (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    att TEXT,
    cod TEXT NOT NULL,
    description TEXT NOT NULL,
    ust_normal INTEGER NOT NULL,
    ust_adapt INTEGER,
    ust_new_tech INTEGER,
    unit_measurement TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS sprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    project_id INT NOT NULL,
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
    tarefa_candidata INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME,
    deleted_at DATETIME,
    FOREIGN KEY(project_id) REFERENCES projects(id),
    FOREIGN KEY(catalog_id) REFERENCES catalog(id),
    FOREIGN KEY(task_id) REFERENCES tasks(id),
    FOREIGN KEY(sprint_id) REFERENCES sprints(id)
);
