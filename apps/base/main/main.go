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
	"errors"
	"strconv"
	"strings"
)

type App struct {
	Name   string
	Label  string
	Sts_id int
}

type File struct {
	Name   string
	Sts_id int //for Table app

	Apps   []*App
	Expand bool
}

func FindInArray(arr []string, name string) int {
	for i, it := range arr {
		if it == name {
			return i
		}
	}
	return -1
}

func FindSelectedFile() *File {

	if store.SelectedFile >= 0 && store.SelectedFile < len(store.Files) {
		return store.Files[store.SelectedFile]
	}
	store.SelectedFile = -1
	return nil
}

func FindSelectedApp() *App {
	file := FindSelectedFile()
	if file == nil {
		return nil
	}

	if store.SelectedApp >= 0 && store.SelectedApp < len(file.Apps) {
		return file.Apps[store.SelectedApp]
	}
	store.SelectedApp = -1
	return nil
}

func (file *File) FindAppName(name string) *App {
	for _, it := range file.Apps {
		if it.Name == name {
			return it
		}
	}
	return nil
}

func FindFile(name string) *File {
	for _, f := range store.Files {
		if f.Name == name {
			return f
		}
	}
	return nil
}

type Storage struct {
	Files []*File

	SelectedFile int
	SelectedApp  int

	SearchFiles string
	SearchApp   string

	createFile string
}
type Translations struct {
	SAVE            string
	SETTINGS        string
	ZOOM            string
	WINDOW_MODE     string
	FULLSCREEN_MODE string
	ABOUT           string
	QUIT            string
	SEARCH          string

	COPYRIGHT string
	WARRANTY  string

	DATE_FORMAT      string
	DATE_FORMAT_EU   string
	DATE_FORMAT_US   string
	DATE_FORMAT_ISO  string
	DATE_FORMAT_TEXT string

	THEME       string
	THEME_OCEAN string
	THEME_RED   string
	THEME_BLUE  string
	THEME_GREEN string
	THEME_GREY  string

	DPI        string
	SHOW_STATS string
	SHOW_GRID  string
	LANGUAGES  string

	NAME        string
	REMOVE      string
	RENAME      string
	CREATE_FILE string

	ALREADY_EXISTS string
	EMPTY_FIELD    string
	INVALID_NAME   string

	IN_USE string

	ADD_APP   string
	CREATE_DB string
}

// https://en.wikipedia.org/wiki/List_of_ISO_639-1_codes
const g_langs = "|English|Chinese(中文)|Hindi(हिंदी)|Spanish(Español)|Russian(Руштина)|Czech(Česky)"

var g_lang_codes = []string{"", "en", "zh", "hi", "es", "ru", "cs"}

func FindLangCode(lng string) int {
	for ii, cd := range g_lang_codes {
		if cd == lng {
			return ii
		}
	}
	return 0
}

func Settings() {
	SA_ColMax(1, 12)
	SA_ColMax(2, 1)

	y := 0

	SA_Text(trns.SETTINGS).Align(1).Show(1, 0, 1, 1)
	y++

	//languages
	{
		SA_Text(trns.LANGUAGES).Align(1).Show(1, y, 1, 1)
		y++

		langs := strings.Split(SA_Info("languages"), "/") //read

		for i, lng := range langs {

			lang_id := FindLangCode(lng)

			SA_DivStart(1, y, 1, 1)
			{
				SA_ColMax(2, 100)
				changed := false

				SA_Text(strconv.Itoa(i+1)+".").Align(1).Show(0, 0, 1, 1)

				SA_DivStart(1, 0, 1, 1)
				{
					SA_Div_SetDrag("lang", uint64(i))
					src, pos, done := SA_Div_IsDrop("lang", true, false, false)
					if done {
						SA_MoveElement(&langs, &langs, int(src), i, pos)
						changed = true
					}
					SA_Image(SA_ResourceBuildAssetPath("", "reorder.png")).Margin(0.17).Show(0, 0, 1, 1)
				}
				SA_DivEnd()

				if SA_Combo(&lang_id, g_langs).Align(0).Search(true).Show(2, 0, 1, 1) {
					langs[i] = g_lang_codes[lang_id]
					changed = true
				}

				if SA_Button("X").Enable(len(langs) > 1 || i > 0).Margin(0.06).Alpha(0.5).Show(3, 0, 1, 1).click {
					langs = append(langs[:i], langs[i+1:]...)
					changed = true
				}

				if changed {
					ll := ""
					for _, lng := range langs {
						ll += lng + "/"
					}
					SA_InfoSet("languages", strings.TrimSuffix(ll, "/"))

					SA_DivEnd() //!
					break
				}
			}
			SA_DivEnd()
			i++
			y++
		}

		SA_DivStart(1, y, 1, 1)
		if SA_Button("+").Margin(0.1).Alpha(0.5).Show(0, 0, 1, 1).click {
			SA_InfoSet("languages", SA_Info("languages")+"/")
		}
		y++
		SA_DivEnd()

		y++ //space
	}

	date := int(SA_InfoFloat("date"))
	if SA_Combo(&date, trns.DATE_FORMAT_EU+"|"+trns.DATE_FORMAT_US+"|"+trns.DATE_FORMAT_ISO+"|"+trns.DATE_FORMAT_TEXT).Align(0).Search(true).ShowDescription(1, y, 1, 2, trns.DATE_FORMAT, 0, 0) {
		SA_InfoSetFloat("date", float64(date))
	}
	y += 3

	theme := int(SA_InfoFloat("theme"))
	if SA_Combo(&theme, trns.THEME_OCEAN+"|"+trns.THEME_RED+"|"+trns.THEME_BLUE+"|"+trns.THEME_GREEN+"|"+trns.THEME_GREY).Align(0).Search(true).ShowDescription(1, y, 1, 2, trns.THEME, 0, 0) {
		SA_InfoSetFloat("theme", float64(theme))
	}
	y += 3

	dpi := strconv.Itoa(int(SA_InfoFloat("dpi")))
	if SA_Editbox(&dpi).ShowDescription(1, y, 1, 2, trns.DPI, 4, 0).finished {
		dpiV, err := strconv.Atoi(dpi)
		if err == nil {
			SA_InfoSetFloat("dpi", float64(dpiV))
		}
	}
	y += 2

	{
		stats := false
		if SA_InfoFloat("stats") > 0 {
			stats = true
		}
		if SA_Checkbox(&stats, trns.SHOW_STATS).Show(1, y, 1, 1) {
			statsV := 0
			if stats {
				statsV = 1
			}
			SA_InfoSetFloat("stats", float64(statsV))
		}
	}
	y++

	{
		grid := false
		if SA_InfoFloat("grid") > 0 {
			grid = true
		}
		if SA_Checkbox(&grid, trns.SHOW_GRID).Show(1, y, 1, 1) {
			gridV := 0
			if grid {
				gridV = 1
			}
			SA_InfoSetFloat("grid", float64(gridV))
		}
	}
	y++

}

func About() {
	SA_ColMax(0, 15)
	SA_Row(1, 3)

	SA_Text(trns.ABOUT).Align(1).Show(0, 0, 1, 1)

	SA_Image(SA_ResourceBuildAssetPath("", "logo.png")).Margin(0.3).InverseColor(true).Show(0, 1, 1, 1)

	SA_Text("v0.1").Align(1).Show(0, 2, 1, 1)

	SA_Button("www.skyalt.com").Alpha(1).Url("https://www.skyalt.com").Show(0, 3, 1, 1)
	SA_Button("github.com/milansuk/skyalt/").Alpha(1).Url("https://github.com/milansuk/skyalt/").Show(0, 4, 1, 1)

	SA_Text(trns.COPYRIGHT).Align(1).Show(0, 5, 1, 1)
	SA_Text(trns.WARRANTY).Align(1).Show(0, 6, 1, 1)
}

func Menu() {
	SA_ColMax(0, 8)
	SA_Row(1, 0.3)
	SA_Row(3, 0.3)
	SA_Row(5, 0.3)
	SA_Row(7, 0.3)
	SA_Row(9, 0.3)

	//save
	if SA_Button(trns.SAVE).Alpha(1).Align(0).Show(0, 0, 1, 1).click {
		SA_InfoSetFloat("save", 1)
		SA_DialogClose()
	}

	SA_RowSpacer(0, 1, 1, 1)

	//settings
	if SA_Button(trns.SETTINGS).Alpha(1).Align(0).Show(0, 2, 1, 1).click {
		SA_DialogClose()
		SA_DialogOpen("Settings", 0)
	}

	SA_RowSpacer(0, 3, 1, 1)

	//zoom
	SA_DivStart(0, 4, 1, 1)
	{
		SA_ColMax(0, 100)
		SA_ColMax(2, 2)

		SA_Text(trns.ZOOM).Show(0, 0, 1, 1)

		dpi := SA_InfoFloat("dpi")
		dpi_default := SA_InfoFloat("dpi_default")
		if SA_Button("+").Margin(0.1).Alpha(0.5).Show(1, 0, 1, 1).click {
			SA_InfoSetFloat("dpi", dpi+3)
		}
		dpiV := int(dpi / dpi_default * 100)
		SA_Text(strconv.Itoa(dpiV)+"%").Align(1).Show(2, 0, 1, 1)
		if SA_Button("-").Margin(0.1).Alpha(0.5).Show(3, 0, 1, 1).click {
			SA_InfoSetFloat("dpi", dpi-3)
		}
	}
	SA_DivEnd()

	SA_RowSpacer(0, 5, 1, 1)

	//window/fullscreen switch
	{
		isFullscreen := SA_InfoFloat("fullscreen")
		ff := trns.WINDOW_MODE
		if isFullscreen == 0 {
			ff = trns.FULLSCREEN_MODE
		}
		if SA_Button(ff).Alpha(1).Align(0).Show(0, 6, 1, 1).click {
			if isFullscreen > 0 {
				isFullscreen = 0
			} else {
				isFullscreen = 1
			}
			SA_InfoSetFloat("fullscreen", isFullscreen)
		}
	}

	SA_RowSpacer(0, 7, 1, 1)

	if SA_Button(trns.ABOUT).Alpha(1).Align(0).Show(0, 8, 1, 1).click {
		SA_DialogClose()
		SA_DialogOpen("About", 0)
	}

	SA_RowSpacer(0, 9, 1, 1)

	if SA_Button(trns.QUIT).Alpha(1).Align(0).Show(0, 10, 1, 1).click {
		SA_InfoSetFloat("exit", 1)
		SA_DialogClose()
	}

}

func Apps(file *File, file_i int) {
	SA_ColMax(0, 7)

	y := 0
	SA_Editbox(&store.SearchApp).TempToValue(true).Ghost(trns.SEARCH).Show(0, 0, 1, 1)
	y++

	apps := strings.Split(SA_Info("apps"), "/") //read
	for _, app := range apps {

		if len(store.SearchApp) > 0 {
			if !strings.Contains(strings.ToLower(app), strings.ToLower(store.SearchApp)) {
				continue
			}
		}

		nm := app
		if file.FindAppName(app) != nil {
			nm += "(" + trns.IN_USE + ")"
		}

		if SA_Button(nm).Alpha(1).Show(0, y, 1, 1).click {
			sts_id := int(SA_InfoFloat("sts_uid"))
			file.Apps = append(file.Apps, &App{Name: app, Label: app, Sts_id: sts_id})
			file.Expand = true
			store.SelectedFile = file_i
			store.SelectedApp = len(file.Apps) - 1
			store.SearchApp = ""
			SA_DialogClose()
		}
		y++
	}

}

func ProjectFiles() {
	files := strings.Split(SA_Info("files"), "/") //read
	apps := strings.Split(SA_Info("apps"), "/")   //read

	//add
	for _, nm := range files {
		if FindFile(nm) == nil {
			store.Files = append(store.Files, &File{Name: nm, Sts_id: int(SA_InfoFloat("sts_uid")), Expand: true})
			store.SelectedFile = len(store.Files) - 1
		}
	}
	//remove
	for i := len(store.Files) - 1; i >= 0; i-- {
		f := store.Files[i]
		if FindInArray(files, f.Name) < 0 {
			store.Files = append(store.Files[:i], store.Files[i+1:]...) //remove
		}
	}

	//remove apps
	for _, f := range store.Files {

		for i := len(f.Apps) - 1; i >= 0; i-- {
			if FindInArray(apps, f.Apps[i].Name) < 0 {
				f.Apps = append(f.Apps[:i], f.Apps[i+1:]...) //remove
			}
		}
	}

	//check selected
	FindSelectedApp()
	/*f := FindFile(store.SelectedFile)
	if f != nil {
		if f.FindApp(store.SelectedApp_sts_id) == nil {
			store.SelectedApp_sts_id = -1
		}
	} else {
		store.SelectedFile = ""
		store.SelectedApp_sts_id = -1
	}*/
}

func Files() {

	ProjectFiles()

	SA_ColMax(0, 100)
	y := 0
	for file_i, file := range store.Files {

		if len(store.SearchFiles) > 0 {
			if !strings.Contains(strings.ToLower(file.Name), strings.ToLower(store.SearchFiles)) {
				continue
			}
		}

		SA_DivStart(0, y, 1, 1)
		{
			SA_Col(0, 0.8)
			SA_ColMax(1, 100)

			isSelected := (file_i == store.SelectedFile && store.SelectedApp < 0)
			if isSelected {
				SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeCd().Aprox(SA_ThemeWhite(), 0.8), 0)
			}

			if len(file.Apps) == 0 {
				file.Expand = false
			}
			iconName := "tree_hide.png"
			if !file.Expand {
				iconName = "tree_show.png"
			}
			if SA_Button("").Alpha(1).Enable(len(file.Apps) > 0).Icon(SA_ResourceBuildAssetPath("", iconName)).MarginIcon(0.15).Show(0, 0, 1, 1).click {
				file.Expand = !file.Expand
			}

			//drop app on file
			SA_DivStart(1, 0, 1, 1)
			{
				src, _, done := SA_Div_IsDrop("app", false, false, true)
				if done {
					src_file_i := uint32(src >> 32)
					src_app_i := uint32(src)

					backup := store.Files[src_file_i].Apps[src_app_i]
					//remove
					store.Files[src_file_i].Apps = append(store.Files[src_file_i].Apps[:src_app_i], store.Files[src_file_i].Apps[src_app_i+1:]...)
					//add
					file.Apps = append(file.Apps, backup)
					file.Expand = true
				}
			}
			SA_DivEnd()

			//name
			SA_DivStart(1, 0, 1, 1)
			{
				SA_ColMax(0, 100)
				if SA_Button(file.Name).Alpha(1).Align(0).Highlight(isSelected).Title("id: "+strconv.Itoa(file.Sts_id)).Show(0, 0, 1, 1).click {
					store.SelectedFile = file_i
					store.SelectedApp = -1
				}
				SA_Div_SetDrag("file", uint64(file_i))
				src, pos, done := SA_Div_IsDrop("file", true, false, false)
				if done {
					SA_MoveElement(&store.Files, &store.Files, int(src), file_i, pos)
				}
			}
			SA_DivEnd()

			//add app
			if SA_Button("+").Title(trns.ADD_APP).Margin(0.15).Alpha(1).Border(true).BackCd(SA_ThemeCd().Aprox(SA_ThemeWhite(), 0.6)).Show(2, 0, 1, 1).click {
				SA_DialogOpen("apps_"+file.Name, 1)
			}
			if SA_DialogStart("apps_" + file.Name) {
				Apps(file, file_i)
				SA_DialogEnd()
			}

			//context
			if SA_Button("").Alpha(1).Icon(SA_ResourceBuildAssetPath("", "context.png")).MarginIcon(0.3).Show(3, 0, 1, 1).click {
				SA_DialogOpen("fileContext_"+file.Name, 1)
			}

			//renameFile := false
			//removeFileConfirm := false
			if SA_DialogStart("fileContext_" + file.Name) {
				SA_ColMax(0, 5)

				if SA_Button(trns.RENAME).Alpha(1).Align(0).Show(0, 0, 1, 1).click {
					SA_DialogOpen("RenameFile_"+file.Name, 1)
				}

				if SA_Button(trns.REMOVE).Alpha(1).Align(0).Show(0, 1, 1, 1).click {
					SA_DialogOpen("RemoveFileConfirm_"+file.Name, 1)
				}

				SA_DialogEnd()
			}

			if SA_DialogStart("RenameFile_" + file.Name) {

				SA_ColMax(0, 7)

				newName := file.Name
				if SA_Editbox(&newName).Error(nil).Show(0, 0, 1, 1).finished {
					if SA_InfoSet("rename_file", file.Name+"/"+newName) {
						file.Name = newName
					}
					SA_DialogClose()
				}

				SA_DialogEnd()
			}

			if SA_DialogStart("RemoveFileConfirm_" + file.Name) {
				if SA_DialogConfirm() {
					if store.SelectedFile == file_i {
						store.SelectedFile = -1
						store.SelectedApp = -1
					}
					SA_InfoSet("remove_file", file.Name)
				}
				SA_DialogEnd() //pokud je pod ním context, tak nastaví context, né root ...
			}
		}
		SA_DivEnd()

		y++

		//apps
		if file.Expand {
			for app_i, app := range file.Apps {

				SA_DivStart(0, y, 1, 1)
				{
					SA_Col(0, 1)
					SA_ColMax(1, 100)

					isSelected := (file_i == store.SelectedFile && app_i == store.SelectedApp)
					if isSelected {
						SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeCd().Aprox(SA_ThemeWhite(), 0.8), 0)
					}

					//name
					SA_DivStart(1, 0, 1, 1)
					{
						SA_ColMax(0, 100)
						if SA_Button(app.Label).Alpha(1).Align(0).Highlight(isSelected).Title("app: "+app.Name+", id: "+strconv.Itoa(app.Sts_id)).Show(0, 0, 1, 1).click {
							store.SelectedFile = file_i
							store.SelectedApp = app_i
						}

						id := (uint64(file_i) << uint64(32)) | uint64(app_i)
						SA_Div_SetDrag("app", id)
						src, pos, done := SA_Div_IsDrop("app", true, false, false)
						if done {
							src_file_i := uint32(src >> 32)
							src_app_i := uint32(src)
							SA_MoveElement(&store.Files[src_file_i].Apps, &file.Apps, int(src_app_i), app_i, pos)
						}
					}
					SA_DivEnd()

					//context
					if SA_Button("").Alpha(1).Icon(SA_ResourceBuildAssetPath("", "context.png")).MarginIcon(0.3).Show(2, 0, 1, 1).click {
						SA_DialogOpen("appContext_"+file.Name+"_"+strconv.Itoa(app.Sts_id), 1)
					}

					//renameApp := false
					//removeAppConfirm := false
					if SA_DialogStart("appContext_" + file.Name + "_" + strconv.Itoa(app.Sts_id)) {
						SA_ColMax(0, 5)

						if SA_Button(trns.RENAME).Alpha(1).Align(0).Show(0, 0, 1, 1).click {
							SA_DialogOpen("RenameApp_"+file.Name+"_"+strconv.Itoa(app.Sts_id), 1)
						}

						if SA_Button(trns.REMOVE).Alpha(1).Align(0).Show(0, 1, 1, 1).click {
							SA_DialogOpen("RemoveAppConfirm_"+file.Name+"_"+strconv.Itoa(app.Sts_id), 1)

						}
						SA_DialogEnd()
					}

					if SA_DialogStart("RenameApp_" + file.Name + "_" + strconv.Itoa(app.Sts_id)) {
						SA_ColMax(0, 7)
						backupLabel := app.Label
						if SA_Editbox(&app.Label).Show(0, 0, 1, 1).finished {
							if len(app.Label) == 0 {
								app.Label = backupLabel
							}
							SA_DialogClose()
						}
						SA_DialogEnd()
					}

					if SA_DialogStart("RemoveAppConfirm_" + file.Name + "_" + strconv.Itoa(app.Sts_id)) {
						if SA_DialogConfirm() {
							if store.SelectedFile == file_i && store.SelectedApp == app_i {
								store.SelectedApp = -1
							}

							file.Apps = append(file.Apps[:app_i], file.Apps[app_i+1:]...) //remove

							SA_DialogEnd() //!
							break
						}
						SA_DialogEnd()
					}

					y++

				}
				SA_DivEnd()
			}
		}
	}

	//new database
	SA_DivStart(0, y, 1, 1)
	{
		if SA_Button("+").Title(trns.CREATE_DB).Margin(0.1).Alpha(0.5).Show(0, 0, 1, 1).click {
			SA_DialogOpen("newFile", 1)
		}
		if SA_DialogStart("newFile") {

			SA_ColMax(0, 9)
			err := CheckFileName(store.createFile, FindFile(store.createFile) != nil)

			SA_Editbox(&store.createFile).Error(err).TempToValue(true).ShowDescription(0, 0, 1, 1, trns.NAME, 2, 0)

			if SA_Button(trns.CREATE_FILE).Enable(err == nil).Show(0, 1, 1, 1).click {
				SA_InfoSet("new_file", store.createFile)
				SA_DialogClose()
			}

			SA_DialogEnd()
		}
	}
	SA_DivEnd()
}

func CheckFileName(name string, alreadyExist bool) error {

	empty := len(name) == 0

	name = strings.ToLower(name)

	var err error
	if alreadyExist {
		err = errors.New(trns.ALREADY_EXISTS)
	} else if empty {
		err = errors.New(trns.EMPTY_FIELD)
	}

	return err
}

//export render
func render() uint32 {

	SA_Col(0, 4.5) //min
	SA_ColResize(0, 7)
	SA_ColMax(1, 100)
	SA_RowMax(1, 100)

	SA_DivStart(0, 0, 1, 1)
	{
		SA_Col(0, 2)
		SA_ColMax(1, 100)

		//Menu + dialogs
		if SA_Button("").Alpha(1).Icon(SA_ResourceBuildAssetPath("", "logo_small.png")).IconInverseColor(true).Margin(0.1).Show(0, 0, 1, 1).click {
			SA_DialogOpen("Menu", 1)
		}
		if SA_DialogStart("Menu") {
			Menu()
			SA_DialogEnd()
		}

		if SA_DialogStart("Settings") {
			Settings()
			SA_DialogEnd()
		}
		if SA_DialogStart("About") {
			About()
			SA_DialogEnd()
		}

		//Search
		SA_Editbox(&store.SearchFiles).TempToValue(true).Margin(0.1).Ghost(trns.SEARCH).HighlightEdit(len(store.SearchFiles) > 0).Show(1, 0, 1, 1)

	}
	SA_DivEnd()

	SA_DivStart(0, 1, 1, 1)
	Files()
	SA_DivEnd()

	file := FindSelectedFile()
	app := FindSelectedApp()
	if app != nil {
		SA_DivStartName(1, 0, 1, 2, strconv.Itoa(app.Sts_id)+"_"+strconv.Itoa(file.Sts_id))

		//file := FindFile(store.SelectedFile)
		//if file != nil {
		//app := file.FindApp(store.SelectedApp_sts_id)
		//if app != nil {
		SA_RenderApp(app.Name, file.Name, app.Sts_id)
		//}
		//}
		SA_DivEnd()
	} else if file != nil {
		SA_DivStartName(1, 0, 1, 2, "_tables_"+strconv.Itoa(file.Sts_id))
		//file := FindFile(store.SelectedFile)
		//if file != nil {
		SA_RenderApp("db", file.Name, file.Sts_id)
		//}
		SA_DivEnd()
	}

	return 0
}

func open(buff []byte) bool {

	//init
	store.SelectedFile = -1
	store.SelectedApp = -1

	return false //default json
}
func save() ([]byte, bool) {
	return nil, false //default json
}
func debug() (int, int, string) {
	return -1, 0, "main" //0=base
}
