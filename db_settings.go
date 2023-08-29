/*
Copyright 2023 Milan Suk

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this db except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DbSettings struct {
	root *Root
	db   *sql.DB

	max_sts_uid int
}

func NewDbSettings(root *Root) (*DbSettings, error) {
	var sts DbSettings
	sts.root = root

	var err error
	sts.db, err = sql.Open("sqlite3", "file:"+sts.GetPath()+"?&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("Open(%s) failed: %w", sts.GetPath(), err)
	}

	_, err = sts.db.Exec("CREATE TABLE IF NOT EXISTS settings(id INT, asset TEXT, content BLOB);")
	if err != nil {
		return nil, fmt.Errorf("Exec(%s) failed: %w", sts.GetPath(), err)
	}

	{
		rows, err := sts.db.Query("SELECT MAX(id) FROM settings")
		if err != nil {
			return nil, fmt.Errorf("Query(%s) failed: %w", sts.GetPath(), err)
		}

		if rows.Next() {
			rows.Scan(&sts.max_sts_uid)

			//fails when 0 rows ...
			//err := rows.Scan(&sts.max_sts_uid)
			//if err != nil {
			//	return nil, fmt.Errorf("Scan(%s) failed: %w", sts.GetPath(), err)
			//}
		}
	}
	return &sts, nil
}

func (sts *DbSettings) Destroy() error {
	return sts.db.Close()
}

func DbSettings_GetName() string {
	return "settings.sqlite"
}

func (sts *DbSettings) GetPath() string {
	return sts.root.folderDatabases + "/" + DbSettings_GetName()
}

func (sts *DbSettings) AddSts_uid() int {
	sts.max_sts_uid++
	return sts.max_sts_uid
}

func (sts *DbSettings) Add(id int, asset string) (int, error) {
	sts.max_sts_uid = OsMax(sts.max_sts_uid, id)

	res, err := sts.db.Exec("INSERT INTO settings(id, asset) VALUES(?, ?);", id, asset)
	if err != nil {
		return -1, fmt.Errorf("Exec(%s) failed: %w", sts.GetPath(), err)
	}

	rowid, err := res.LastInsertId()
	if err != nil {
		return -1, fmt.Errorf("LastInsertId(%s) failed: %w", sts.GetPath(), err)
	}

	return int(rowid), nil
}

func (sts *DbSettings) Find(id int, asset string) (int, error) {
	rows, err := sts.db.Query("SELECT rowid FROM settings WHERE id=? AND asset=?", id, asset)
	if err != nil {
		return -1, fmt.Errorf("Query(%s) failed: %w", sts.GetPath(), err)
	}

	rowid := -1 //not found
	if rows.Next() {
		err := rows.Scan(&rowid)
		if err != nil {
			return -1, fmt.Errorf("Scan(%s) failed: %w", sts.GetPath(), err)
		}
	}
	return rowid, nil
}

func (sts *DbSettings) FindOrAdd(id int, asset string) (int, error) {
	rowid, err := sts.Find(id, asset)
	if err != nil {
		return -1, fmt.Errorf("Find() failed: %w", err)
	}

	if rowid < 0 {
		rowid, err = sts.Add(id, asset)
		if err != nil {
			return -1, fmt.Errorf("Add() failed: %w", err)
		}
	}

	return rowid, nil
}

func (sts *DbSettings) GetContent(rowid int) ([]byte, error) {

	rows := sts.db.QueryRow("SELECT content FROM settings WHERE rowid=?", rowid)

	var content []byte
	err := rows.Scan(&content)
	if err != nil {
		return nil, fmt.Errorf("Scan(%s) failed: %w", sts.GetPath(), err)
	}
	return content, nil
}

func (sts *DbSettings) SetContent(rowid int, content []byte) error {

	_, err := sts.db.Exec("UPDATE settings SET content=? WHERE rowid=?;", content, rowid)
	if err != nil {
		return fmt.Errorf("Exec(%s) failed: %w", sts.GetPath(), err)
	}
	return nil
}

func (sts *DbSettings) Remove(rowid int) error {
	_, err := sts.db.Exec("DELETE FROM settings WHERE rowid=?;", rowid)
	if err != nil {
		return fmt.Errorf("Exec(%s) failed: %w", sts.GetPath(), err)
	}
	return nil
}

func (sts *DbSettings) Rename(rowid int, file, app, name string) error {

	_, err := sts.db.Exec("UPDATE settings SET file=?, app=?, name=? WHERE rowid=?;", file, app, name, rowid)
	if err != nil {
		return fmt.Errorf("Exec(%s) failed: %w", sts.GetPath(), err)
	}
	return nil
}
