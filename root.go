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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
)

type Root struct {
	ctx             context.Context
	folderApps      string
	folderDatabases string
	folderDevice    string

	cacheDir      string
	cache         wazero.CompilationCache
	runtimeConfig wazero.RuntimeConfig

	apps []*App
	dbs  map[string]*Db

	dbsList  string
	appsList string

	last_ticks int64

	level *LayoutLevel
	touch LayerTouch
	tile  LayerTile

	ui *Ui

	baseApp string
	baseDb  string

	infoLayout RS_LScroll
	//stack      Stack
	stack *LayoutLevel

	fonts *Fonts

	ui_info Info
	vm_info Info

	editbox_history VmTextHistoryArray

	server *DebugServer

	settings *DbSettings

	exit bool
	save bool
}

func NewRoot(debugPORT int, folderApps string, folderDbs string, folderDevice string, ctx context.Context) (*Root, error) {
	var root Root
	var err error
	root.ctx = ctx

	root.fonts = NewFonts()
	root.dbs = make(map[string]*Db)

	root.folderApps = folderApps
	root.folderDatabases = folderDbs
	root.folderDevice = folderDevice

	os.Mkdir(folderApps, 0700)
	os.Mkdir(folderDbs, 0700)
	os.Mkdir(folderDevice, 0700)

	if !OsFolderExists(folderApps) {
		return nil, fmt.Errorf("Folder(%s) not exist", folderApps)
	}
	if !OsFolderExists(folderDbs) {
		return nil, fmt.Errorf("Folder(%s) not exist", folderDbs)
	}
	if !OsFolderExists(folderDevice) {
		return nil, fmt.Errorf("Folder(%s) not exist", folderDevice)
	}

	root.settings, err = NewDbSettings(&root)
	if err != nil {
		return nil, fmt.Errorf("NewDbSettings() failed: %w", err)
	}

	// init wasm
	root.cacheDir, err = os.MkdirTemp("", "wasm_cache")
	if err != nil {
		return nil, fmt.Errorf("MkdirTemp() failed: %w", err)
	}
	root.cache, err = wazero.NewCompilationCacheWithDir(root.cacheDir)
	if err != nil {
		return nil, fmt.Errorf("NewCompilationCacheWithDir() failed: %w", err)
	}
	root.runtimeConfig = wazero.NewRuntimeConfig().WithCompilationCache(root.cache)

	iniPath, scrollPath, err := root.GetSettingsPaths()
	if err != nil {
		return nil, fmt.Errorf("GetSettingsPaths() failed: %w", err)
	}

	root.ui, err = NewUi(iniPath)
	if err != nil {
		return nil, fmt.Errorf("NewUi() failed: %w", err)
	}

	//scroll
	err = root.infoLayout.Open(scrollPath)
	if err != nil {
		return nil, fmt.Errorf("Open() failed: %w", err)
	}

	root.level = NewLayoutLevel(nil, "", OsV4{}, nil, &root.infoLayout, root.ui)

	root.baseApp = "base"
	root.baseDb = "base"

	root.updateDbsList()
	root.updateAppsList()

	root.server, err = NewDebugServer(debugPORT)
	if err != nil {
		return nil, fmt.Errorf("NewDebugServer() failed: %w", err)
	}

	return &root, nil
}
func (root *Root) Destroy() {

	for _, app := range root.apps {
		app.Destroy()
	}

	if root.server != nil {
		root.server.Destroy()
	}

	for nm, db := range root.dbs {
		err := db.Destroy()
		if err != nil {
			fmt.Printf("db(%s).Destroy() failed: %v\n", nm, err)
		}
	}

	root.fonts.Destroy()

	root.cache.Close(root.ctx)
	os.RemoveAll(root.cacheDir)

	//save settings
	{
		iniPath, scrollPath, err := root.GetSettingsPaths()
		if err != nil {
			fmt.Printf("GetSettingsPaths() failed: %v\n", err)
		}

		err = root.ui.io.Save(iniPath)
		if err != nil {
			fmt.Printf("Open() failed: %v\n", err)
		}

		root.level.div.Save(&root.infoLayout)
		err = root.infoLayout.Save(scrollPath)
		if err != nil {
			fmt.Printf("Open() failed: %v\n", err)
		}
	}

	root.settings.Destroy()

	root.ui.Destroy() //also save ini.json
}

func (root *Root) SetLevel(act *LayoutLevel) {
	root.stack = act
	act.stack = act.div

	act.stack.data.touch_enabled = act.IsTop()
}

func (root *Root) ReloadTranslations() {
	for _, app := range root.apps {
		app.ReloadTranslations()
	}
}

func (root *Root) GetSettingsPaths() (string, string, error) {

	dev, err := os.Hostname()
	if err != nil {
		return "", "", fmt.Errorf("Hostname() failed: %w", err)
	}
	return root.folderDevice + "/" + dev + "_ini.json", "device/" + dev + "_scroll.json", nil
}

func (root *Root) FindApp(appName string, dbName string, sts_id int) *App {
	for _, app := range root.apps {
		if app.name == appName && (len(dbName) == 0 || app.db_name == dbName) && (sts_id < 0 || app.sts_id == sts_id) {
			return app
		}
	}
	return nil
}

func (root *Root) AddApp(appName string, dbName string, sts_id int) (*App, error) {
	//find
	app := root.FindApp(appName, dbName, sts_id)
	if app != nil {
		return app, nil //ok
	}

	//add db
	_, err := root.AddDb(dbName)
	if err != nil {
		return nil, err
	}

	//add
	app, err = NewApp(root, appName, dbName, sts_id)
	if err != nil {
		return nil, err
	}
	root.apps = append(root.apps, app)
	return app, nil
}

func (root *Root) AddDb(name string) (*Db, error) {

	//finds
	db, found := root.dbs[name]
	if found {
		return db, nil
	}

	//adds
	var err error
	db, err = NewDb(root, name)
	if err != nil {
		return nil, err
	}

	root.dbs[name] = db
	return db, nil
}

func (root *Root) CreateDb(name string) bool {

	newPath := root.folderDatabases + "/" + name + ".sqlite"
	if OsFileExists(newPath) {
		fmt.Printf("newPath(%s) already exist\n", newPath)
		return false
	}

	f, err := os.Create(newPath)
	if err != nil {
		fmt.Printf("Create(%s) failed: %v\n", newPath, name)
		return false
	}

	err = f.Close()
	if err != nil {
		fmt.Printf("Close(%s) failed: %v\n", newPath, name)
		return false
	}

	root.updateDbsList()
	return true
}

func (root *Root) RenameDb(name string, newName string) bool {

	if strings.ContainsRune(newName, '/') || strings.ContainsRune(newName, '\\') {
		fmt.Printf("newName(%s) has invalid character\n", name)
		return false
	}

	path := root.folderDatabases + "/" + name + ".sqlite"
	newPath := root.folderDatabases + "/" + newName + ".sqlite"
	if OsFileExists(newPath) {
		fmt.Printf("newPath(%s) already exist\n", newPath)
		return false
	}

	//finds
	db, found := root.dbs[name]
	if found {
		//close
		err := db.Destroy()
		if err != nil {
			fmt.Printf("db(%s).Destroy() failed: %v\n", name, err)
		}
		delete(root.dbs, name)
	}

	//rename file
	err := OsFileRename(path, newPath)
	if err != nil {
		fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
	}
	if OsFileExists(path + "-shm") {
		err = OsFileRename(path+"-shm", newPath+"-shm")
		if err != nil {
			fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
		}
	}
	if OsFileExists(path + "-wal") {
		err = OsFileRename(path+"-wal", newPath+"-wal")
		if err != nil {
			fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
		}
	}

	root.updateDbsList()
	return true
}

func (root *Root) RemoveDb(name string) bool {

	//finds
	db, found := root.dbs[name]
	if found {
		//close
		err := db.Destroy()
		if err != nil {
			fmt.Printf("db(%s).Destroy() failed: %v\n", name, err)
		}
		delete(root.dbs, name)
	}

	//delete file
	path := root.folderDatabases + "/" + name + ".sqlite"
	err := OsFileRemove(path)
	if err != nil {
		fmt.Printf("OsFileRemove(%s) failed: %v\n", path, err)
	}
	if OsFileExists(path + "-shm") {
		err = OsFileRemove(path + "-shm")
		if err != nil {
			fmt.Printf("OsFileRemove(%s-shm) failed: %v\n", path, err)
		}
	}
	if OsFileExists(path + "-wal") {
		err = OsFileRemove(path + "-wal")
		if err != nil {
			fmt.Printf("OsFileRemove(%s-wal) failed: %v\n", path, err)
		}
	}

	root.updateDbsList()
	return true
}

func (root *Root) CommitDbs() {
	for _, db := range root.dbs {
		if db.tx == nil {
			continue
		}

		err := db.Commit()
		if err != nil {
			fmt.Printf("Commit() failed: %v\n", err)
		}
	}
}

func (root *Root) Render() {

	winRect, _ := root.ui.GetScreenCoord()
	root.level.div.canvas = winRect
	root.level.div.crop = winRect

	// close all levels
	if root.ui.io.keys.shift && root.ui.io.keys.esc {
		root.touch.Reset()
		root.level.next.Delete()
		root.ui.io.keys.esc = false
	}

	ist, err := root.AddApp(root.baseApp, root.baseDb, 0)
	if err != nil {
		fmt.Printf("AddApp(%s).Db(%s) failed: %v\n", root.baseApp, root.baseDb, err)
		return
	}

	root.SetLevel(root.level)

	st := root.stack
	//background
	st.buff.Reset(st.stack.canvas)
	st.buff.AddCrop(st.stack.canvas)
	st.buff.AddRect(st.stack.canvas, OsCd_white(), 0)

	ist.Render(true)

	//close un-used
	//draw
	act := root.level
	act.use = true //base level is always use
	for act != nil {
		if !act.use || act.close {
			act.Delete()
		}
		act.use = false
		act = act.next
	}

	//layout maintenance
	act = root.level
	for act != nil {
		act.div.Maintenance(&root.infoLayout)
		act = act.next
	}

	//draw
	act = root.level
	for act != nil {
		if act.buff != nil {
			act.buff.Draw()
		}
		act = act.next
	}
}

func (root *Root) Tick() (bool, error) {

	if time.Now().UnixMilli() > root.last_ticks+2000 {
		for _, app := range root.apps {
			app.Tick()
		}
		root.last_ticks = time.Now().UnixMilli()

		root.updateDbsList()
		root.updateAppsList()
	}

	run, err := root.ui.UpdateIO()
	if err != nil {
		return false, fmt.Errorf("UpdateIO() failed: %w", err)
	}

	//tile
	{
		if root.tile.NeedsRedrawFromSleep(root.ui.io.touch.pos) {
			root.ui.ResendInput()
		}
		root.tile.NextTick()
	}

	if root.ui.NeedRedraw() {

		stUiTicks := OsTicks()
		root.ui.StartRender()
		stVmTicks := OsTicks()

		if root.ui.io.touch.start {
			root.touch.Reset()
		}

		root.Render()

		if root.ui.io.touch.end {
			root.touch.Reset()
			root.ui.io.drag.group = ""
		}

		// tile - redraw If mouse is over tile
		if root.tile.IsActive(root.ui.io.touch.pos) {
			err := root.ui.RenderTile(root.tile.text, root.tile.coord, root.tile.cd, root.fonts.Get(0))
			if err != nil {
				fmt.Printf("RenderTile() failed: %v\n", err)
			}
		}

		// show fps
		if root.ui.io.ini.Stats {
			root.ui.RenderInfoStats(&root.ui_info, &root.vm_info, root.fonts.Get(0))
		}

		root.vm_info.Update(int(OsTicks() - stVmTicks))
		root.ui.EndRender()
		root.ui_info.Update(int(OsTicks() - stUiTicks))

		if root.save {
			for _, app := range root.apps {
				app.SaveData()
			}
			root.save = false
		}

	} else {
		time.Sleep(10 * time.Millisecond)
	}

	root.CommitDbs()

	return (run && !root.exit), err
}

func (root *Root) updateDbsList() {

	dir, err := os.ReadDir(root.folderDatabases)
	if err != nil {
		fmt.Printf("ReadDir() failed: %v\n", err)
		return
	}

	root.dbsList = ""
	for _, file := range dir {
		if !file.IsDir() {
			ext := filepath.Ext(file.Name())
			if strings.EqualFold(ext, ".sqlite") && file.Name() != DbSettings_GetName() && file.Name() != "base.sqlite" {
				root.dbsList += strings.TrimSuffix(file.Name(), ext) + "/"
			}
		}
	}
	root.dbsList = strings.TrimSuffix(root.dbsList, "/") //remove '/' at the end
}

func (root *Root) updateAppsList() {

	dir, err := os.ReadDir(root.folderApps)
	if err != nil {
		fmt.Printf("ReadDir() failed: %v\n", err)
		return
	}

	root.appsList = ""
	for _, file := range dir {
		if file.IsDir() && file.Name() != "base" {
			root.appsList += file.Name() + "/"
		}
	}
	root.appsList = strings.TrimSuffix(root.appsList, "/") //remove '/' at the end
}
