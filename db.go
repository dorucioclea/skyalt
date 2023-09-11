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
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"fmt"
	"math"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

type DbCache struct {
	query_hash  int64
	query       string
	result_rows [][]byte
}

// 100% copied from sdk.go
func _argsToArray(data []byte, arg interface{}) []byte {

	switch v := arg.(type) {

	case bool:
		data = append(data, TpI64)
		if v {
			data = binary.LittleEndian.AppendUint64(data, 1)
		} else {
			data = binary.LittleEndian.AppendUint64(data, 0)
		}
	case byte:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))
	case int:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))
	case uint:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))

	case int16:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))
	case uint16:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))

	case int32:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))
	case int64:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))

	case uint32:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))
	case uint64:
		data = append(data, TpI64)
		data = binary.LittleEndian.AppendUint64(data, uint64(v))

	case float32:
		data = append(data, TpF32)
		data = binary.LittleEndian.AppendUint64(data, uint64(math.Float32bits(v)))

	case float64:
		data = append(data, TpF64)
		data = binary.LittleEndian.AppendUint64(data, uint64(math.Float64bits(v)))

	case []byte:
		data = append(data, TpBytes)
		data = binary.LittleEndian.AppendUint64(data, uint64(len(v)))
		data = append(data, v...)
	case string:
		data = append(data, TpBytes)
		data = binary.LittleEndian.AppendUint64(data, uint64(len(v)))
		data = append(data, v...)

	case nil:
		data = append(data, TpBytes)
		data = binary.LittleEndian.AppendUint64(data, 0)
	}

	return data
}

func NewDbCache(query string, db *sql.DB) (*DbCache, error) {
	var cache DbCache

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("Query(%s) failed: %w", query, err)
	}

	cache.query = query
	{
		h := sha256.New()
		h.Write([]byte(query))
		cache.query_hash = int64(binary.LittleEndian.Uint64(h.Sum(nil)))
	}

	// out fields
	colTypes, err := rows.ColumnTypes()
	if err != nil {
		return nil, fmt.Errorf("ColumnTypes() failed: %w", err)
	}

	values := make([]interface{}, len(colTypes))
	scanCallArgs := make([]interface{}, len(colTypes))
	for i := range colTypes {
		scanCallArgs[i] = &values[i]
	}

	for rows.Next() {
		//reset
		for i := range values {
			values[i] = nil
		}
		err := rows.Scan(scanCallArgs...)
		if err != nil {
			return nil, fmt.Errorf("Scan() failed: %w", err)
		}

		var row []byte
		for _, v := range values {
			row = _argsToArray(row, v)
		}
		cache.result_rows = append(cache.result_rows, row)
	}

	return &cache, nil
}

type Db struct {
	root *Root
	name string

	db *sql.DB
	tx *sql.Tx

	cache []*DbCache

	lastChange int
}

func NewDb(root *Root, name string) (*Db, error) {
	var db Db
	db.root = root
	db.name = name

	var err error
	db.db, err = sql.Open("sqlite3", "file:"+db.GetPath()+"?&_journal_mode=WAL")
	if err != nil {
		return nil, fmt.Errorf("Open(%s) failed: %w", db.GetPath(), err)
	}

	db.UpdateTime()

	return &db, nil
}

func (db *Db) Destroy() error {
	return db.db.Close()
}

func (db *Db) Begin() (*sql.Tx, error) {
	if db.tx == nil {
		var err error
		db.tx, err = db.db.Begin()
		if err != nil {
			return nil, fmt.Errorf("Begin(%s) failed: %w", db.GetPath(), err)
		}
	}
	return db.tx, nil
}

func (db *Db) GetPath() string {
	return db.root.folderDatabases + "/" + db.name + ".sqlite"
}

func (db *Db) Commit() error {
	err := db.tx.Commit()
	db.tx = nil

	//reset queries
	db.cache = nil

	return err
}

func (db *Db) ReOpen() error {
	err := db.db.Close()
	if err != nil {
		return fmt.Errorf("Close(%s) failed: %w", db.GetPath(), err)
	}

	db.db, err = sql.Open("sqlite3", db.GetPath())
	if err != nil {
		return fmt.Errorf("Open(%s) failed: %w", db.GetPath(), err)
	}
	return nil
}

func (db *Db) UpdateTime() bool {

	info, err := os.Stat(db.GetPath())
	if os.IsNotExist(err) {
		return false
	}

	diff := info.ModTime().Unix() != int64(db.lastChange)

	db.lastChange = int(info.ModTime().Unix())
	return diff

}

func (db *Db) FindCache(query_hash int64) *DbCache {

	//find
	for _, it := range db.cache {
		if it.query_hash == query_hash {
			return it
		}
	}
	return nil
}

func (db *Db) AddCache(query string) (*DbCache, error) {

	//find
	for _, it := range db.cache {
		if it.query == query {
			return it, nil
		}
	}

	//add
	cache, err := NewDbCache(query, db.db)
	if err != nil {
		return nil, fmt.Errorf("NewDbCache(%s) failed: %w", db.GetPath(), err)
	}

	db.cache = append(db.cache, cache)
	return cache, nil
}

func (db *Db) Write(query string, params ...any) (sql.Result, error) {

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	res, err := tx.Exec(query, params...)
	if err != nil {
		return nil, fmt.Errorf("query(%s) failed: %w", query, err)
	}

	return res, nil
}
