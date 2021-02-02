package stmt

const CreateTables = `

CREATE TABLE IF NOT EXISTS note
(
  id            text    PRIMARY KEY,
  type          text    NOT NULL,
  title         text    NOT NULL,
  size          int     NOT NULL,
  deleted       int     NOT NULL,
  remind_at     text    NOT NULL,
  created_at    text    NOT NULL,
  updated_at    text    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_note_remind ON note(remind_at);
CREATE INDEX IF NOT EXISTS idx_note_create ON note(created_at);
CREATE INDEX IF NOT EXISTS idx_note_update ON note(updated_at);

CREATE TABLE IF NOT EXISTS tag
(
  id            text    PRIMARY KEY,
  name          text    NOT NULL UNIQUE,
  created_at    text    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tag_create ON tag(created_at);

CREATE TABLE IF NOT EXISTS note_tag
(
  note_id   text    REFERENCES note(ID) ON DELETE CASCADE,
  tag_id    text    REFERENCES tag(ID)  ON DELETE CASCADE,
  UNIQUE (note_id, tag_id)
);

CREATE TABLE IF NOT EXISTS patch
(
  id      text    PRIMARY KEY,
  diff    text    NOT NULL
);

CREATE TABLE IF NOT EXISTS note_patch
(
  note_id     text    REFERENCES note(ID) ON DELETE CASCADE,
  patch_id    text    REFERENCES patch(ID)  ON DELETE CASCADE,
  UNIQUE (note_id, patch_id)
);

CREATE TABLE IF NOT EXISTS file
(
  id            text    PRIMARY KEY,
  name          text    NOT NULL,
  size          int     NOT NULL,
  type          text    NOT NULL,
  checksum      text    NOT NULL UNIQUE,
  deleted       int     NOT NULL,
  created_at    text    NOT NULL,
  updated_at    text    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_file_create ON file(created_at);
CREATE INDEX IF NOT EXISTS idx_file_update ON file(updated_at);

CREATE TABLE IF NOT EXISTS note_file
(
  note_id    text    REFERENCES note(ID) ON DELETE CASCADE,
  file_id    text    REFERENCES file(ID)  ON DELETE CASCADE,
  UNIQUE (note_id, file_id)
);

CREATE TABLE IF NOT EXISTS taggroup
(
  id            text    PRIMARY KEY,
  tags          blob    NOT NULL UNIQUE,
  protected     int     NOT NULL,
  created_at    text    NOT NULL,
  updated_at    text    NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_taggroup_create ON taggroup(created_at);
CREATE INDEX IF NOT EXISTS idx_taggroup_update ON taggroup(updated_at);

CREATE TABLE IF NOT EXISTS metadata
(
  name         text    NOT NULL UNIQUE,
  int_value    int     DEFAULT NULL,
  text_value   text    DEFAULT NULL
);
`

const InsertIntValue = `INSERT INTO metadata (name, int_value) VALUES (?, ?);`
const GetIntValue = `SELECT int_value FROM metadata WHERE name=?;`
const UpdateIntValue = `UPDATE metadata SET int_value=? WHERE name=?;`

const InsertTextValue = `INSERT INTO metadata (name, text_value) VALUES (?, ?);`
const GetTextValue = `SELECT text_value FROM metadata WHERE name=?;`
const UpdateTextValue = `UPDATE metadata SET text_value=? WHERE name=?;`

const GetNote = `SELECT * FROM note WHERE id=?;`
const GetNoteSize = `SELECT size FROM note WHERE id=?;`
const GetAllNoteIDs = `SELECT id FROM note`
const GetNotes = `SELECT * FROM note WHERE deleted=0 ORDER BY updated_at;`
const GetDeletedNotes = `SELECT * FROM note WHERE deleted>0 ORDER BY updated_at;`
const SearchNoteTitle = `SELECT * FROM note WHERE title LIKE ?`
const GetNotesByTagID = `SELECT note_id FROM note_tag WHERE tag_id=?;`
const InsertNote = `INSERT INTO note (
    id, type, title, size, deleted, remind_at, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
const UpdateTitleSizeNow = `UPDATE note SET title=?, size=?, updated_at=?
    WHERE id=?;`
const SetNoteDeleted = `UPDATE note SET deleted=? WHERE id=?;`
const DeleteNote = `DELETE FROM note WHERE id=?;`
const ChangeNoteType = `UPDATE note SET type=? WHERE id=?;`
const SetTypeTitle = `UPDATE note SET type=?, title=? WHERE id=?;`

const GetTag = `SELECT * FROM tag WHERE id=?;`
const GetTagByName = `SELECT * FROM tag WHERE name=?;`
const GetTagID = `SELECT id FROM tag WHERE name=?;`
const InsertTag = `INSERT INTO tag (id, name, created_at) VALUES (?, ?, ?);`
const RenameTag = `UPDATE tag SET name=? WHERE id=?; `
const DeleteTag = `DELETE FROM tag WHERE id=?;`
const InsertNoteTag = `INSERT INTO note_tag (note_id, tag_id) VALUES (?, ?);`
const DeleteTagsByNote = `DELETE FROM note_tag WHERE note_id=?;`
const DeleteTags = `DELETE FROM note_tag WHERE note_id=? and tag_id=?;`

const InsertPatch = `INSERT INTO patch (id, diff) VALUES (?, ?);`
const InsertNotePatch = `INSERT INTO note_patch (note_id, patch_id) VALUES (?, ?);`

const InsertFile = `INSERT INTO file (
    id, name, size, type, checksum, deleted, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?, ?, ?, ?);`
const InsertNoteFile = `INSERT INTO note_file (note_id, file_id) VALUES (?, ?);`

const GetTagGroup = `SELECT * FROM taggroup WHERE id=?;`
const AllTagGroups = `SELECT * FROM taggroup ORDER BY updated_at;`
const GetTagGroupID = `SELECT id FROM taggroup WHERE tags=?;`
const TagGroupCount = `SELECT count(*) FROM taggroup`
const LastTagGroup = `SELECT id FROM taggroup WHERE protected=0
    ORDER BY updated_at LIMIT 1;`
const DeleteTagGroup = `DELETE FROM taggroup WHERE id=?;`
const InsertTagGroup = `INSERT INTO taggroup (
    id, tags, protected, created_at, updated_at)
    VALUES (?, ?, ?, ?, ?);`
const UpdateTagGroupNow = `UPDATE taggroup SET updated_at=? WHERE id=?;`
const SetTagGroupProtected = `UPDATE taggroup SET protected=? WHERE id=?;`

const GetNotesByTagName = `SELECT note_tag.note_id FROM tag
    INNER JOIN note_tag ON tag.id = note_tag.tag_id
    WHERE tag.name=?;`

const AllTagsByName = `SELECT tag.id, tag.name, tag.created_at,
    (SELECT count(*) FROM note_tag WHERE note_tag.tag_id = tag.id)
    FROM tag ORDER BY tag.name;`

const AllTagsByDate = `SELECT tag.id, tag.name, tag.created_at,
    (SELECT count(*) FROM note_tag WHERE note_tag.tag_id = tag.id)
    FROM tag ORDER BY tag.created_at;`

const GetTagsByNote = `SELECT tag.id, tag.name FROM note
    INNER JOIN note_tag ON note.id = note_tag.note_id
    INNER JOIN tag ON note_tag.tag_id = tag.id
    WHERE note.id=?;`

const GetPatchesByNote = `SELECT patch.diff FROM note
    INNER JOIN note_patch ON note.id = note_patch.note_id
    INNER JOIN patch ON note_patch.patch_id = patch.id
    WHERE note.id=?;`
