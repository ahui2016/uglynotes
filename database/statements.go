package database

const CreateTables = `

CREATE TABLE IF NOT EXISTS note
(
  id		text	PRIMARY KEY,
  type		text	NOT NULL,
  title		text	NOT NULL,
  size		int	NOT NULL,
  deleted	int	NOT NULL,
  remind_at	text	NOT NULL,
  created_at	text	NOT NULL,
  updated_at	text	NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_note_remind ON note(remind_at);
CREATE INDEX IF NOT EXISTS idx_note_create ON note(created_at);
CREATE INDEX IF NOT EXISTS idx_note_update ON note(updated_at);

CREATE TABLE IF NOT EXISTS tag
(
  id		text	PRIMARY KEY,
  name		text	NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS note_tag
(
  note_id	text	REFERENCES note(ID) ON DELETE CASCADE,
  tag_id	text	REFERENCES tag(ID)  ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS patch
(
  id	text	PRIMARY KEY,
  diff	text	NOT NULL
);

CREATE TABLE IF NOT EXISTS note_patch
(
  note_id	text	REFERENCES note(ID) ON DELETE CASCADE,
  patch_id	text	REFERENCES tag(ID)  ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS file
(
  id		text	PRIMARY KEY,
  name		text	NOT NULL,
  size		int	NOT NULL,
  type		text	NOT NULL,
  checksum	text	NOT NULL UNIQUE,
  deleted	int	NOT NULL,
  created_at	text	NOT NULL,
  updated_at	text	NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_file_create ON file(created_at);
CREATE INDEX IF NOT EXISTS idx_file_update ON file(updated_at);

CREATE TABLE IF NOT EXISTS note_file
(
  note_id	text	REFERENCES note(ID) ON DELETE CASCADE,
  file_id	text	REFERENCES tag(ID)  ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS taggroup
(
  id		text	PRIMARY KEY,
  tags		blob	NOT NULL UNIQUE,
  protected	int	NOT NULL,
  created_at	text	NOT NULL,
  updated_at	text	NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_taggroup_create ON taggroup(created_at);
CREATE INDEX IF NOT EXISTS idx_taggroup_update ON taggroup(updated_at);
`
