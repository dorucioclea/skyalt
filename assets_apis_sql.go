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

import "fmt"

func (asset *Asset) _getDb(dbName string) (*Db, error) {
	var err error
	var db *Db
	if len(dbName) == 0 {
		db, err = asset.app.root.AddDb(asset.app.db_name)
	} else {
		db, err = asset.app.root.AddDb(dbName)
	}

	return db, err
}

func (asset *Asset) sql_write(dbName string, query string) (int64, error) {

	db, err := asset._getDb(dbName)
	if err != nil {
		return -1, err
	}

	db.tx, err = db.Begin()
	if err != nil {
		return -1, fmt.Errorf("Begin(%s) failed: %w", db.GetPath(), err)
	}

	res, err := db.tx.Exec(query)
	if err != nil {
		return -1, fmt.Errorf("Exec(%s) for query(%s) failed: %w", db.GetPath(), query, err)
	}

	aff, err := res.RowsAffected()
	if err != nil {
		return -1, fmt.Errorf("RowsAffected(%s) failed: %w", db.GetPath(), err)
	}

	if aff <= 0 {
		return 0, nil
	}
	return 1, nil
}

func (asset *Asset) _sa_sql_write(dbMem uint64, queryMem uint64) int64 {

	db, err := asset.ptrToString(dbMem)
	if asset.AddLogErr(err) {
		return -1
	}
	query, err := asset.ptrToString(queryMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.sql_write(db, query)
	asset.AddLogErr(err)
	return ret
}

func (asset *Asset) sql_read(dbName string, query string) (int64, error) {

	db, err := asset._getDb(dbName)
	if db == nil {
		return -1, err
	}
	cache, err := db.AddCache(query)
	if err != nil {
		return -1, err
	}

	return cache.query_hash, nil
}
func (asset *Asset) _sa_sql_read(dbMem uint64, queryMem uint64) int64 {

	db, err := asset.ptrToString(dbMem)
	if asset.AddLogErr(err) {
		return -1
	}
	query, err := asset.ptrToString(queryMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.sql_read(db, query)
	asset.AddLogErr(err)
	return ret

}

func (asset *Asset) sql_readRowCount(dbName string, query string, queryHash int64) (int64, error) {

	db, err := asset._getDb(dbName)
	if db == nil {
		return -1, err
	}

	cache := db.FindCache(queryHash)

	if cache == nil {
		var err error
		cache, err = db.AddCache(query)
		if err != nil {
			return -1, err
		}
	}

	return int64(len(cache.result_rows)), nil
}

func (asset *Asset) sql_readRowLen(dbName string, query string, queryHash int64, row_i uint64) (int64, error) {

	db, err := asset._getDb(dbName)
	if db == nil {
		return -1, err
	}

	cache := db.FindCache(queryHash)

	if cache == nil {
		var err error
		cache, err = db.AddCache(query)
		if err != nil {
			return -1, err
		}
	}

	if row_i < uint64(len(cache.result_rows)) {
		return int64(len(cache.result_rows[row_i])), nil
	}

	return 0, nil //no more rows
}

func (asset *Asset) _sa_sql_readRowCount(dbMem uint64, queryMem uint64, queryHash int64) int64 {

	db, err := asset.ptrToString(dbMem)
	if asset.AddLogErr(err) {
		return -1
	}
	query, err := asset.ptrToString(queryMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.sql_readRowCount(db, query, queryHash)
	asset.AddLogErr(err)
	return ret
}

func (asset *Asset) _sa_sql_readRowLen(dbMem uint64, queryMem uint64, queryHash int64, row_i uint64) int64 {

	db, err := asset.ptrToString(dbMem)
	if asset.AddLogErr(err) {
		return -1
	}
	query, err := asset.ptrToString(queryMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.sql_readRowLen(db, query, queryHash, row_i)
	asset.AddLogErr(err)
	return ret
}

func (asset *Asset) sql_readRow(dbName string, query string, queryHash int64, row_i uint64) ([]byte, int64, error) {

	db, err := asset._getDb(dbName)
	if db == nil {
		return nil, -1, err
	}

	cache := db.FindCache(queryHash)
	if cache == nil {
		var err error
		cache, err = db.AddCache(query)
		if err != nil {
			return nil, -1, err
		}
	}

	if row_i < uint64(len(cache.result_rows)) {
		return cache.result_rows[row_i], 1, nil
	}

	return nil, 0, nil //no more rows
}

func (asset *Asset) _sa_sql_readRow(dbMem uint64, queryMem uint64, queryHash int64, row_i uint64, resultMem uint64) int64 {

	db, err := asset.ptrToString(dbMem)
	if asset.AddLogErr(err) {
		return -1
	}
	query, err := asset.ptrToString(queryMem)
	if asset.AddLogErr(err) {
		return -1
	}

	dst, ret, err := asset.sql_readRow(db, query, queryHash, row_i)
	asset.AddLogErr(err)

	if ret > 0 {
		err := asset.bytesToPtr(dst, resultMem)
		asset.AddLogErr(err)
		if err != nil {
			return -1
		}
	}

	return ret
}
