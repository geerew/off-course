-- +goose Up

--- Course
CREATE TABLE courses (
	id           TEXT PRIMARY KEY NOT NULL,
	title        TEXT NOT NULL,
	path         TEXT UNIQUE NOT NULL,
	card_path    TEXT,
	available    BOOLEAN NOT NULL DEFAULT FALSE,
	duration     INTEGER NOT NULL DEFAULT 0,
	initial_scan BOOLEAN NOT NULL DEFAULT FALSE,
	maintenance  BOOLEAN NOT NULL DEFAULT FALSE,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Progress of courses
CREATE TABLE courses_progress (
	id           TEXT PRIMARY KEY NOT NULL,
	course_id    TEXT NOT NULL,
	user_id 	 TEXT NOT NULL ,
	started      BOOLEAN NOT NULL DEFAULT FALSE,
	started_at   TEXT,
	percent      INTEGER NOT NULL DEFAULT 0,
	completed_at TEXT,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
	---
	UNIQUE(course_id, user_id)
);

--- Asset groups
CREATE TABLE asset_groups (
	id          	 TEXT PRIMARY KEY NOT NULL,
	course_id   	 TEXT NOT NULL,
	title       	 TEXT NOT NULL,
	prefix      	 INTEGER NOT NULL,
	module      	 TEXT,
	description_path TEXT,
	description_type TEXT,
	created_at       TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at       TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
	UNIQUE(course_id, prefix, module)
);

--- Assets
CREATE TABLE assets (
	id             TEXT PRIMARY KEY NOT NULL,
	course_id      TEXT NOT NULL,
	asset_group_id TEXT NOT NULL,
	title          TEXT NOT NULL,
	prefix         INTEGER NOT NULL,
	sub_prefix     INTEGER,
	sub_title	   TEXT,
	module         TEXT,
	type           TEXT NOT NULL,
	path           TEXT UNIQUE NOT NULL,
	file_size      INTEGER NOT NULL DEFAULT 0,
	mod_time       TEXT NOT NULL DEFAULT '',
	hash	       TEXT NOT NULL,
	created_at     TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at     TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
	FOREIGN KEY (asset_group_id) REFERENCES asset_groups (id) ON DELETE CASCADE
);

--- Asset metadata
CREATE TABLE asset_video_metadata (
	id         TEXT PRIMARY KEY NOT NULL,
	asset_id   TEXT NOT NULL UNIQUE,
	duration   INTEGER NOT NULL DEFAULT 0,
	width      INTEGER NOT NULL DEFAULT 0,
	height     INTEGER NOT NULL DEFAULT 0,
	resolution TEXT NOT NULL DEFAULT '',
	codec      TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	--
	FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
);

--- Progress of assets
CREATE TABLE assets_progress (
	id           TEXT PRIMARY KEY NOT NULL,
	asset_id     TEXT NOT NULL,
	user_id      TEXT NOT NULL,
	video_pos    INTEGER NOT NULL DEFAULT 0,
	completed	 BOOLEAN NOT NULL DEFAULT FALSE,
	completed_at TEXT,
	created_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at   TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (asset_id) REFERENCES assets (id) ON DELETE CASCADE,
	FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
	---
	UNIQUE(asset_id, user_id)
);

--- Attachments (related to asset groups)
CREATE TABLE attachments (
	id             TEXT PRIMARY KEY NOT NULL,
	asset_group_id TEXT NOT NULL,
	title          TEXT NOT NULL,
	path           TEXT UNIQUE NOT NULL,
	created_at     TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at     TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (asset_group_id) REFERENCES asset_groups (id) ON DELETE CASCADE
);

--- Scan jobs
CREATE TABLE scans (
	id         TEXT PRIMARY KEY NOT NULL,
	course_id  TEXT UNIQUE NOT NULL,
    status     TEXT NOT NULL DEFAULT 'waiting',
	created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE
);

--- Tags
CREATE TABLE tags (
	id         TEXT PRIMARY KEY NOT NULL,
    tag        TEXT NOT NULL COLLATE NOCASE UNIQUE,
	created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Course tags (join table)
CREATE TABLE courses_tags (
	id         TEXT PRIMARY KEY NOT NULL,
	tag_id     TEXT NOT NULL,
	course_id  TEXT NOT NULL,
	created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
	---
	FOREIGN KEY (tag_id) REFERENCES tags (id) ON DELETE CASCADE,
	FOREIGN KEY (course_id) REFERENCES courses (id) ON DELETE CASCADE,
	---
	CONSTRAINT unique_course_tag UNIQUE (tag_id, course_id)
);

--- Parameters (for application settings)
CREATE TABLE params (
    id         TEXT PRIMARY KEY NOT NULL,
    key        TEXT UNIQUE NOT NULL,
    value      TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Users
CREATE TABLE users (
    id            TEXT PRIMARY KEY NOT NULL,
    username      TEXT UNIQUE NOT NULL COLLATE NOCASE,
	display_name  TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role          TEXT NOT NULL CHECK(role IN ('admin', 'user')),
    created_at    TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW')),
    updated_at    TEXT NOT NULL DEFAULT (STRFTIME('%Y-%m-%d %H:%M:%f', 'NOW'))
);

--- Sessions
CREATE TABLE sessions (
    id      TEXT PRIMARY KEY NOT NULL,
	data    BLOB NOT NULL,
	expires BIGINT NOT NULL,
	user_id TEXT NOT NULL DEFAULT ''
);

-- Groups by course, then ordered by prefix+module
CREATE INDEX idx_asset_groups_course_prefix_module
  ON asset_groups(course_id, prefix, module);

-- Assets: WHERE asset_group_id = ? ORDER BY prefix, sub_prefix
CREATE INDEX idx_assets_group_prefix_sub
  ON assets(asset_group_id, prefix, sub_prefix);

-- Attachments: WHERE asset_group_id = ? ORDER BY title
CREATE INDEX idx_attachments_group_title
  ON attachments(asset_group_id, title);

-- Asset progress with user
CREATE INDEX idx_asset_progress_asset_user
  ON assets_progress(asset_id, user_id);

-- Video metadata
CREATE UNIQUE INDEX idx_video_metadata_asset
  ON asset_video_metadata(asset_id);

-- Filter assets by course quickly
CREATE INDEX IF NOT EXISTS idx_assets_course ON assets(course_id);

-- Probe progress rows by (asset_id, user_id)
CREATE INDEX IF NOT EXISTS idx_asset_progress_asset_user 
	ON assets_progress(asset_id, user_id);

-- Sessions
CREATE INDEX idx_sessions_expires ON sessions(expires);
CREATE INDEX idx_sessions_user    ON sessions(user_id);