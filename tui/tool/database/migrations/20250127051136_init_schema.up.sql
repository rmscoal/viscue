CREATE TABLE IF NOT EXISTS categories(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name VARCHAR NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS passwords(
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER,
    name VARCHAR NOT NULL,
    email VARCHAR,
    username VARCHAR,
    password VARCHAR NOT NULL,

    FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX idx_name_per_category ON passwords (name, category_id);

CREATE TABLE IF NOT EXISTS configurations(
    key VARCHAR PRIMARY KEY,
    value TEXT
);
