package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gpuwatch/internal/types"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct{ *sql.DB }

func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?_foreign_keys=on", path))
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &DB{db}, nil
}

func (db *DB) Close() error { return db.DB.Close() }

func migrate(db *sql.DB) error {
	stmts := []string{
		`PRAGMA journal_mode=WAL;`,
		`CREATE TABLE IF NOT EXISTS snapshots (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			ts INTEGER NOT NULL
		);`,
		`CREATE TABLE IF NOT EXISTS gpu_stats (
			snapshot_id INTEGER NOT NULL,
			gpu_index INTEGER,
			name TEXT, uuid TEXT,
			util_gpu REAL, util_mem REAL,
			mem_used_mb REAL, mem_total_mb REAL,
			temp_c REAL, power_w REAL, power_limit_w REAL,
			FOREIGN KEY(snapshot_id) REFERENCES snapshots(id) ON DELETE CASCADE
		);`,
		`CREATE TABLE IF NOT EXISTS proc_stats (
			snapshot_id INTEGER NOT NULL,
			gpu_uuid TEXT,
			pid INTEGER,
			process_name TEXT,
			used_mem_mb REAL,
			user TEXT,
			FOREIGN KEY(snapshot_id) REFERENCES snapshots(id) ON DELETE CASCADE
		);`,
		`CREATE INDEX IF NOT EXISTS idx_snapshots_ts ON snapshots(ts);`,
		`CREATE INDEX IF NOT EXISTS idx_proc_snapshot ON proc_stats(snapshot_id);`,
		`CREATE INDEX IF NOT EXISTS idx_gpu_snapshot ON gpu_stats(snapshot_id);`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

// SaveSnapshot persists a full snapshot and returns its ID.
func (db *DB) SaveSnapshot(s types.Snapshot) (int64, error) {
	tx, err := db.Begin()
	if err != nil { return 0, err }
	defer func(){ if err != nil { _ = tx.Rollback() } }()

	res, err := tx.Exec(`INSERT INTO snapshots(ts) VALUES (?)`, s.TS.Unix())
	if err != nil { return 0, err }
	id, err := res.LastInsertId()
	if err != nil { return 0, err }

	for _, g := range s.GPUs {
		_, err = tx.Exec(`INSERT INTO gpu_stats(snapshot_id,gpu_index,name,uuid,util_gpu,util_mem,mem_used_mb,mem_total_mb,temp_c,power_w,power_limit_w)
			VALUES(?,?,?,?,?,?,?,?,?,?,?)`, id, g.Index, g.Name, g.UUID, g.UtilGPU, g.UtilMem, g.MemUsedMB, g.MemTotalMB, g.TempC, g.PowerDrawW, g.PowerLimitW)
		if err != nil { return 0, err }
	}
	for _, p := range s.Procs {
		_, err = tx.Exec(`INSERT INTO proc_stats(snapshot_id,gpu_uuid,pid,process_name,used_mem_mb,user)
			VALUES(?,?,?,?,?,?)`, id, p.GPUUUID, p.PID, p.ProcessName, p.UsedMemMB, p.User)
		if err != nil { return 0, err }
	}
	if err = tx.Commit(); err != nil { return 0, err }
	return id, nil
}

// SnapshotMeta minimal info for navigation.
type SnapshotMeta struct {
	ID int64
	TS time.Time
}

// ListSnapshotsByDate returns all snapshot metas for a given local date (00:00-24:00).
func (db *DB) ListSnapshotsByDate(day time.Time) ([]SnapshotMeta, error) {
	start := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	end := start.Add(24 * time.Hour)
	rows, err := db.Query(`SELECT id, ts FROM snapshots WHERE ts >= ? AND ts < ? ORDER BY ts ASC`, start.Unix(), end.Unix())
	if err != nil { return nil, err }
	defer rows.Close()
	var out []SnapshotMeta
	for rows.Next() {
		var id int64
		var tsUnix int64
		if err := rows.Scan(&id, &tsUnix); err != nil { return nil, err }
		out = append(out, SnapshotMeta{ID: id, TS: time.Unix(tsUnix, 0).In(day.Location())})
	}
	return out, rows.Err()
}

// LoadSnapshot loads a full snapshot by id.
func (db *DB) LoadSnapshot(id int64) (types.Snapshot, error) {
	var tsUnix int64
	err := db.QueryRow(`SELECT ts FROM snapshots WHERE id=?`, id).Scan(&tsUnix)
	if err != nil { return types.Snapshot{}, err }
	s := types.Snapshot{ID: id, TS: time.Unix(tsUnix, 0)}
	// GPUs
	rows, err := db.Query(`SELECT gpu_index,name,uuid,util_gpu,util_mem,mem_used_mb,mem_total_mb,temp_c,power_w,power_limit_w FROM gpu_stats WHERE snapshot_id=?`, id)
	if err != nil { return types.Snapshot{}, err }
	for rows.Next() {
		var g types.GPU
		if err := rows.Scan(&g.Index,&g.Name,&g.UUID,&g.UtilGPU,&g.UtilMem,&g.MemUsedMB,&g.MemTotalMB,&g.TempC,&g.PowerDrawW,&g.PowerLimitW); err != nil { rows.Close(); return types.Snapshot{}, err }
		s.GPUs = append(s.GPUs, g)
	}
	rows.Close()
	// Procs
	rows, err = db.Query(`SELECT gpu_uuid,pid,process_name,used_mem_mb,user FROM proc_stats WHERE snapshot_id=?`, id)
	if err != nil { return types.Snapshot{}, err }
	for rows.Next() {
		var p types.GPUProcess
		if err := rows.Scan(&p.GPUUUID,&p.PID,&p.ProcessName,&p.UsedMemMB,&p.User); err != nil { rows.Close(); return types.Snapshot{}, err }
		s.Procs = append(s.Procs, p)
	}
	rows.Close()
	return s, nil
}

var ErrNoSnapshots = errors.New("no snapshots")

// LoadLatest returns the latest snapshot or ErrNoSnapshots.
func (db *DB) LoadLatest() (types.Snapshot, error) {
	var id, ts int64
	err := db.QueryRow(`SELECT id, ts FROM snapshots ORDER BY ts DESC, id DESC LIMIT 1`).Scan(&id, &ts)
	if err == sql.ErrNoRows { return types.Snapshot{}, ErrNoSnapshots }
	if err != nil { return types.Snapshot{}, err }
	return db.LoadSnapshot(id)
}
