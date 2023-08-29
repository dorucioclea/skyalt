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
	"fmt"
	"strconv"
	"strings"
)

type Storage struct {
	Tables        []*Table
	SelectedTable int

	//ShowViewPanel bool

	renameTable  string
	createTable  string
	createColumn string

	showFilterDialog bool
}

type Translations struct {
	CREATE_TABLE string
	RENAME       string
	REMOVE       string
	DUPLICATE    string

	ALREADY_EXISTS string
	EMPTY_FIELD    string
	INVALID_NAME   string

	CREATE_VIEW string

	COLUMNS  string
	SHOW_ALL string
	HIDE_ALL string

	FILTER string
	SORT   string

	ENABLE      string
	NAME        string
	ROWS_HEIGHT string

	AND string
	OR  string

	TEXT  string
	FLOAT string

	HIDE string

	ADD_ROW string

	MIN   string
	MAX   string
	AVG   string
	SUM   string
	COUNT string
}

type FilterItem struct {
	Column string
	Op     int
	Value  string
}

func (f *FilterItem) GetOpString() string {

	switch f.Op {
	case 0:
		return "="
	case 1:
		return "!="
	case 2:
		return "<="
	case 3:
		return ">="
	case 4:
		return "<"
	case 5:
		return ">"
	}
	return ""
}
func Filter_getOptions() string {
	return "=|<>|<=|>=|<|>|"
}

type Filter struct {
	Enable bool
	Items  []*FilterItem
	Rel    int
}

func (f *Filter) UpdateColumn(old string, new string) {
	for _, it := range f.Items {
		if it.Column == old {
			it.Column = new
		}
	}
}

func (f *Filter) Add(columnName string, op int) {
	f.Items = append(f.Items, &FilterItem{Column: columnName, Op: op})
}

func (f *Filter) Check() {
	//toto smaže neplatné, ale možná by se měli přejmenovat na ""
}

type SortItem struct {
	Column string
	Az     int
}
type Sort struct {
	Enable bool
	Items  []*SortItem
}

func (s *Sort) UpdateColumn(old string, new string) {
	for _, it := range s.Items {
		if it.Column == old {
			it.Column = new
		}
	}
}

func (s *Sort) Find(columnName string) *SortItem {
	for _, it := range s.Items {
		if it.Column == columnName {
			return it
		}
	}
	return nil
}
func (s *Sort) Add(columnName string, az int) bool {

	if len(columnName) == 0 || s.Find(columnName) == nil {
		s.Items = append(s.Items, &SortItem{Column: columnName, Az: az})
		return true
	}
	return false
}

func (s *Sort) Check() {

}

type Column struct {
	Name   string
	Type   string
	Show   bool
	Resize float64

	Render string //checkbox, etc.

	StatFunc string
}

func (col *Column) isRowId() bool {
	return col.Name == "rowid"
}

type Table struct {
	Name    string
	Columns []*Column

	Filter  Filter
	Sort    Sort
	RowSize int //0=>1row, 1=2rows

	scrollDown bool
}

func (table *Table) UpdateColumn(old string, new string) {
	//for _, w := range table.Views {
	table.Filter.UpdateColumn(old, new)
	table.Sort.UpdateColumn(old, new)
	//}
}

/*func (table *Table) AddView(name string) *View {
	vw := &View{Name: name}
	vw.Filter.Enable = true
	vw.Sort.Enable = true

	// rowid column
	vw.Columns = append(vw.Columns, &Column{Name: "rowid", Type: "INTEGER"})

	table.Views = append(table.Views, vw)
	return vw
}*/

/*func (table *Table) Check() {
	if len(table.Views) == 0 {
		table.AddView("View_0")
	}
	if table.SelectedView >= len(table.Views) {
		table.SelectedView = 0
	}
}*/

func GetDbStructure() []*Table {
	var tables []*Table

	qt := SA_SqlRead("", "SELECT name FROM sqlite_master WHERE type = 'table'")
	var tname string
	for qt.Next(&tname) {

		//table
		table := Table{Name: tname}
		//view := table.AddView("View_0")
		table.Filter.Enable = true
		table.Sort.Enable = true
		// rowid column
		table.Columns = append(table.Columns, &Column{Name: "rowid", Type: "INTEGER"})

		//column
		qc := SA_SqlRead("", "pragma table_info("+tname+");")
		var cid int
		var cname, ctype string
		for qc.Next(&cid, &cname, &ctype) {
			resize := float64(4)
			if cname == "rowid" {
				resize = 1
			}
			table.Columns = append(table.Columns, &Column{Name: cname, Type: ctype, Show: true, Resize: resize})
		}

		tables = append(tables, &table)
	}

	return tables
}

func (table *Table) GetRender(columnName string) string {
	//for _, view := range table.Views {
	for _, cl := range table.Columns {
		if cl.Name == columnName && len(cl.Render) > 0 {
			return cl.Render
		}
	}
	//}
	return ""
}

func (table *Table) SetRender(columnName string, render string) {
	//for _, view := range table.Views {
	for _, cl := range table.Columns {
		if cl.Name == columnName {
			cl.Render = render
		}
	}
	//}
}

func FindTable(tables []*Table, tname string) *Table {
	for _, tb := range tables {
		if tb.Name == tname {
			return tb
		}
	}
	return nil
}

func (table *Table) FindColumn(cname string) *Column {
	for _, cl := range table.Columns {
		if cl.Name == cname {
			return cl
		}
	}
	return nil
}

/*func (table *Table) FindView(wname string) *View {
	for _, vw := range table.Views {
		if vw.Name == wname {
			return vw
		}
	}
	return nil
}*/

func UpdateViews() {

	db := GetDbStructure()

	//add tables
	for _, db_tb := range db {
		if FindTable(store.Tables, db_tb.Name) == nil {
			store.Tables = append(store.Tables, db_tb)
		}
	}

	//add columns
	for _, table := range store.Tables {
		//check
		//table.Check()

		db_tb := FindTable(db, table.Name)
		if db_tb != nil {
			//for _, view := range table.Views {
			for _, db_cl := range db_tb.Columns {
				column := table.FindColumn(db_cl.Name)
				if column == nil {
					column = db_cl
					table.Columns = append(table.Columns, column)
				}
				column.Type = db_cl.Type
				column.Render = table.GetRender(column.Name)
			}
			//}
		}
	}

	//remove tables/Columns
	for ti := len(store.Tables) - 1; ti >= 0; ti-- {
		table := store.Tables[ti]

		db_tb := FindTable(db, table.Name)
		if db_tb == nil {
			store.Tables = append(store.Tables[:ti], store.Tables[ti+1:]...) //remove table
			continue
		}

		for ci := len(table.Columns) - 1; ci >= 0; ci-- {
			column := table.Columns[ci]

			db_cl := db_tb.FindColumn(column.Name)
			if db_cl == nil {
				table.Columns = append(table.Columns[:ci], table.Columns[ci+1:]...) //remove column
				//continue
			}
		}
		//}
	}
	if store.SelectedTable >= len(store.Tables) {
		store.SelectedTable = 0
	}

	//fix Columns
	for _, table := range store.Tables {
		//for _, view := range table.Views {
		for _, column := range table.Columns {
			if column.isRowId() {
				column.Show = true
			}
		}
		//}
	}

	//fix filter/short
	for _, table := range store.Tables {
		//for _, view := range table.Views {
		table.Filter.Check()
		table.Sort.Check()
		//}
	}
}

func DragAndDropTable(dst int) {
	SA_Div_SetDrag("table", dst, false, true, false)
	src, pos, done := SA_Div_IsDrop("table")
	if done {
		selTable := store.Tables[store.SelectedTable]
		SA_MoveElement(store.Tables, src, dst, pos)

		for i, tb := range store.Tables {
			if tb == selTable {
				store.SelectedTable = i
			}
		}
	}
}

func DragAndDropColumn(dst int, table *Table) {
	SA_Div_SetDrag("column", dst, false, true, false)
	src, pos, done := SA_Div_IsDrop("column")
	if done {
		SA_MoveElement(table.Columns, src, dst, pos)
	}
}

/*func DragAndDropView(dst int, table *Table) {
	SA_Div_SetDrag("view", dst, true, false, false)
	src, pos, done := SA_Div_IsDrop("view")
	if done {
		selView := table.Views[table.SelectedView]
		SA_MoveElement(table.Views, src, dst, pos)

		for i, vw := range table.Views {
			if vw == selView {
				table.SelectedView = i
			}
		}
	}
}*/

func TablesList() {
	SA_DivSetInfo("scrollHnarrow", 1)
	SA_DivSetInfo("scrollVshow", 0)

	for x := range store.Tables {
		SA_Col(x, 3)
		SA_ColMax(x, 5)
	}

	for x, table := range store.Tables {
		SA_DivStart(x, 0, 1, 1)
		{
			SA_ColMax(0, 5)

			isSelected := (store.SelectedTable == x)

			openTableMenu := false
			openRenameTable := false
			removeTableConfirm := false
			if SA_Button(table.Name).Alpha(1).Align(1).Highlight(isSelected).Show(0, 0, 1, 1).click {
				store.SelectedTable = x
				if isSelected {
					openTableMenu = true
				}
			}

			DragAndDropTable(x)

			if SA_DialogStart("TableMenu_"+table.Name, 1, openTableMenu) {
				SA_ColMax(0, 5)
				SA_Row(1, 0.3)

				if SA_Button(trns.RENAME).Alpha(1).Show(0, 0, 1, 1).click {
					store.renameTable = table.Name
					SA_DialogClose()
					openRenameTable = true
				}

				//space
				SA_RowSpacer(0, 1, 1, 1)

				if SA_Button(trns.REMOVE).BackCd(SA_ThemeWarning()).Show(0, 2, 1, 1).click {
					SA_DialogClose()
					removeTableConfirm = true
				}

				SA_DialogEnd()
			}

			if SA_DialogStart("RenameTable_"+table.Name, 1, openRenameTable) {
				RenameTable(table)
				SA_DialogEnd()
			}

			if SA_DialogStart("RemoveTableConfirm_"+table.Name, 1, removeTableConfirm) {
				if SA_DialogConfirm(0, 0, 1, 1) {
					SA_SqlWrite("", "DROP TABLE "+table.Name+";")
				}

				SA_DialogEnd()
			}

		}
		SA_DivEnd()
	}

}

func CheckName(name string, alreadyExist bool) error {

	empty := len(name) == 0

	name = strings.ToLower(name)
	invalidName := !empty && (name[0] < 'a' || name[0] > 'z') //first must be a-z

	var err error
	if alreadyExist {
		err = errors.New(trns.ALREADY_EXISTS)
	} else if empty {
		err = errors.New(trns.EMPTY_FIELD)
	} else if invalidName {
		err = errors.New(trns.INVALID_NAME)
	}

	return err
}

func CreateTable() {
	SA_ColMax(0, 9)

	err := CheckName(store.createTable, FindTable(store.Tables, store.createTable) != nil)

	SA_Editbox(&store.createTable).Error(err).TempToValue(true).ShowDescription(0, 0, 1, 1, trns.NAME, 2, 0)

	if SA_Button(trns.CREATE_TABLE).Enable(err == nil).Show(0, 1, 1, 1).click {
		SA_SqlWrite("", "CREATE TABLE "+store.createTable+"(column TEXT DEFAULT '' NOT NULL);")
		SA_DialogClose()
	}
}

func RenameTable(table *Table) {
	SA_ColMax(0, 7)
	SA_ColMax(1, 3)

	err := CheckName(store.renameTable, FindTable(store.Tables, store.renameTable) != nil)

	SA_Editbox(&store.renameTable).Error(err).TempToValue(true).Show(0, 0, 1, 1)

	if SA_Button(trns.RENAME).Enable(err == nil).Show(1, 0, 1, 1).click {
		SA_SqlWrite("", "ALTER TABLE "+table.Name+" RENAME TO "+store.renameTable+";")
		table.Name = store.renameTable
		SA_DialogClose()
	}
}

func TopHeader() {
	SA_ColMax(1, 100)
	SA_Col(2, 2)

	openCreateTable := false
	if SA_Button("+").Align(1).Title(trns.CREATE_TABLE).Show(0, 0, 1, 1).click {
		openCreateTable = true
	}
	if SA_DialogStart("CreateTable", 1, openCreateTable) {
		CreateTable()
		SA_DialogEnd()
	}

	SA_DivStart(1, 0, 1, 1)
	TablesList()
	SA_DivEnd()

	/*if SA_Button(trns.VIEWS).Alpha(1).Highlight(store.ShowViewPanel).Show(2, 0, 1, 1).click {
		store.ShowViewPanel = !store.ShowViewPanel
	}*/
}

func Reorder[T any](x, y, w, h int, group string, id int, array []T) {

	SA_DivStart(x, y, w, h)
	{
		SA_Div_SetDrag(group, id, true, false, false)
		src, pos, done := SA_Div_IsDrop(group)
		if done {
			SA_MoveElement(array, src, id, pos)
		}
		SA_Image(SA_ResourceBuildAssetPath("", "reorder.png")).Margin(0.17).Show(0, 0, 1, 1)
	}
	SA_DivEnd()
}

func TableView(table *Table) {
	//table.Check()

	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	SA_DivStart(0, 0, 1, 1)
	{
		//view := table.Views[table.SelectedView]

		SA_ColMax(0, 5)

		//filter
		SA_Col(1, 0.5)
		SA_ColMax(2, 5)

		//sort
		SA_Col(3, 0.5)
		SA_ColMax(4, 5)

		//rows height
		SA_Col(5, 0.5)
		SA_ColMax(6, 4)

		//SA_Text(view.Name).Show(0, 0, 1, 1)

		hidden := false
		for _, col := range table.Columns {
			if !col.Show {
				hidden = true
			}
		}

		openColumns := false
		openFilter := false
		openSort := false
		if SA_Button(trns.COLUMNS).Alpha(0.5).Highlight(hidden).Show(0, 0, 1, 1).click {
			openColumns = true
		}

		if SA_DialogStart("Columns", 1, openColumns) {

			SA_ColMax(0, 5)
			SA_ColMax(1, 5)
			y := 0
			for i, col := range table.Columns {
				if col.isRowId() {
					continue
				}

				SA_DivStart(0, y, 2, 1)
				{
					SA_ColMax(1, 100)

					Reorder(0, 0, 1, 1, "column2", i, table.Columns)
					SA_Checkbox(&col.Show, col.Name).Show(1, 0, 1, 1)

					y++
				}
				SA_DivEnd()
			}

			if SA_Button(trns.SHOW_ALL).Show(0, y, 1, 1).click {
				for _, col := range table.Columns {
					col.Show = true
				}
			}

			if SA_Button(trns.HIDE_ALL).Show(1, y, 1, 1).click {
				for _, col := range table.Columns {
					if !col.isRowId() {
						col.Show = false
					}
				}
			}

			SA_DialogEnd()
		}

		if SA_Button(trns.FILTER).Alpha(0.5).Highlight(table.Filter.Enable && len(table.Filter.Items) > 0).Show(2, 0, 1, 1).click || store.showFilterDialog {
			store.showFilterDialog = false
			openFilter = true
		}

		if SA_DialogStart("Filter", 1, openFilter) {

			SA_ColMax(0, 2)
			SA_ColMax(1, 6)
			SA_ColMax(2, 4)
			//SA_ColMax(3, 3)

			//enable
			y := 0
			SA_Checkbox(&table.Filter.Enable, trns.ENABLE).Show(0, y, 2, 1)

			//and/or
			SA_Combo(&table.Filter.Rel, trns.AND+"|"+trns.OR).Enable(table.Filter.Enable).Search(true).Show(2, y, 1, 1)
			y++

			for fi, it := range table.Filter.Items {

				SA_DivStart(0, y, 3, 1)
				{
					SA_ColMax(1, 5)
					SA_ColMax(2, 2)
					SA_ColMax(3, 3)

					if table.Filter.Enable {
						Reorder(0, 0, 1, 1, "filter", fi, table.Filter.Items)
					}

					SA_DivStart(1, 0, 1, 1)
					ColumnsCombo(table, &it.Column, table.Filter.Enable)
					SA_DivEnd()

					SA_Combo(&it.Op, Filter_getOptions()).Enable(table.Filter.Enable).Search(true).Show(2, 0, 1, 1)

					SA_Editbox(&it.Value).Enable(table.Filter.Enable).Show(3, 0, 1, 1)

					if SA_Button("X").Enable(table.Filter.Enable).Show(4, 0, 1, 1).click {
						table.Filter.Items = append(table.Filter.Items[:fi], table.Filter.Items[fi+1:]...) //remove
						break
					}
				}
				SA_DivEnd()

				y++
			}

			if SA_Button("+").Enable(table.Filter.Enable).Show(0, y, 1, 1).click {
				table.Filter.Add("", 0)
			}

			SA_DialogEnd()
		}

		if SA_Button(trns.SORT).Alpha(0.5).Highlight(table.Sort.Enable && len(table.Sort.Items) > 0).Show(4, 0, 1, 1).click {
			openSort = true
		}

		if SA_DialogStart("Sort", 1, openSort) {

			SA_ColMax(2, 7)

			y := 0
			SA_Checkbox(&table.Sort.Enable, trns.ENABLE).Show(0, y, 3, 1)
			y++

			for si, it := range table.Sort.Items {

				SA_DivStart(0, y, 3, 1)
				{
					SA_ColMax(1, 5)
					SA_ColMax(2, 2)

					if table.Sort.Enable {
						Reorder(0, 0, 1, 1, "sort", si, table.Sort.Items)
					}

					SA_DivStart(1, 0, 1, 1)
					ColumnsCombo(table, &it.Column, table.Sort.Enable)
					SA_DivEnd()

					SA_Combo(&it.Az, "A -> Z|Z -> A").Enable(table.Sort.Enable).Show(2, 0, 1, 1)

					if SA_Button("X").Enable(table.Sort.Enable).Show(3, 0, 1, 1).click {
						table.Sort.Items = append(table.Sort.Items[:si], table.Sort.Items[si+1:]...) //remove
						break
					}
				}
				SA_DivEnd()

				y++
			}

			if SA_Button("+").Enable(table.Sort.Enable).Show(0, y, 2, 1).click {
				table.Sort.Add("", 0)
			}

			SA_DialogEnd()
		}

		SA_Combo(&table.RowSize, "1|2|3|4").ShowDescription(6, 0, 1, 1, trns.ROWS_HEIGHT, 2.5, 0)

	}
	SA_DivEnd()

	SA_DivStart(0, 1, 1, 1)
	Tablee(table)
	SA_DivEnd()
}

func ColumnsCombo(table *Table, selectedColumn *string, enable bool) {
	SA_ColMax(0, 100)

	pos := -1
	var opts string
	for i, col := range table.Columns {
		opts += col.Name + "|"
		if *selectedColumn == col.Name {
			pos = i
		}
	}
	if len(opts) > 0 {
		opts = opts[:len(opts)-1] //cut last '|'
	}

	var err error
	if pos < 0 {
		err = errors.New("Column not exist")
	}
	if SA_Combo(&pos, opts).Search(true).Error(err).Enable(enable).Show(0, 0, 1, 1) {
		*selectedColumn = table.Columns[pos].Name
	}
}

func ColumnDetail(table *Table, column *Column) {

	SA_ColMax(0, 10)

	SA_DivStart(0, 0, 1, 1)
	{
		SA_ColMax(0, 100)
		SA_ColMax(1, 3)

		origName := column.Name
		if SA_Editbox(&column.Name).ShowDescription(0, 0, 1, 1, trns.NAME, 2, 0).finished {
			SA_SqlWrite("", "ALTER TABLE "+table.Name+" RENAME COLUMN "+origName+" TO "+column.Name+";")

			//update filter/short
			table.UpdateColumn(origName, column.Name)
		}

		if SA_Button(column.Type).Show(1, 0, 1, 1).click {

			//convert types ...
			if column.Type == "REAL" {
				//dialog checkbox, etc. ...
				//table.SetRender(column.Name, "CHECK_BOX")
			}

		}
	}
	SA_DivEnd()

	//sort/filter
	SA_DivStart(0, 2, 1, 2)
	{
		SA_ColMax(0, 100)
		SA_ColMax(1, 100)
		SA_ColMax(2, 100)
		SA_Text(trns.SORT).Show(0, 0, 1, 1)

		//sort
		sort_notUse := table.Sort.Find(column.Name) == nil
		if SA_Button("A -> Z").Enable(sort_notUse).Show(1, 0, 1, 1).click {
			table.Sort.Add(column.Name, 0)
		}
		if SA_Button("Z -> A").Enable(sort_notUse).Show(2, 0, 1, 1).click {
			table.Sort.Add(column.Name, 1)
		}

		//filter
		if SA_Button(trns.FILTER).Align(0).Show(0, 1, 1, 1).click {
			table.Sort.Add(column.Name, 0)

			table.Filter.Add(column.Name, 0)

			store.showFilterDialog = true
			SA_DialogClose()
		}
	}
	SA_DivEnd()

	//properties
	SA_DivStart(0, 5, 1, 3)
	{
		SA_ColMax(0, 100)
		if column.Render == "RATING" {
			//	...Gui.EditboxDesc(0, 0, 1, 1, _value: _column.render.max_stars, maxWidth: 3, asNumber: 1, description:{MAX_STARS})
		}

	}
	SA_DivEnd()

	//hide
	if SA_Button(trns.HIDE).Show(0, 8, 1, 1).click {
		column.Show = false
		SA_DialogClose()
	}

	//remove
	removeColumnConfirm := false
	if SA_Button(trns.REMOVE).BackCd(SA_ThemeWarning()).Show(0, 10, 1, 1).click {
		removeColumnConfirm = true
	}

	if SA_DialogStart("RemoveColumnConfirm", 1, removeColumnConfirm) {
		if SA_DialogConfirm(0, 0, 1, 1) {
			SA_SqlWrite("", "ALTER TABLE "+table.Name+" DROP COLUMN "+column.Name+";")
			SA_DialogClose()
		}
		SA_DialogEnd()
	}

}

func Tablee(table *Table) {

	//table.Check()

	//view := table.Views[table.SelectedView]

	sumWidth := 1.5 //"+"
	for _, col := range table.Columns {
		if col.Show {
			sumWidth += float64(col.Resize)
		}
	}

	SA_Col(0, sumWidth)
	SA_RowMax(1, 100)

	//columns header
	SA_DivStart(0, 0, 1, 1)
	{
		x := 0
		for _, col := range table.Columns {
			if !col.Show {
				continue
			}
			SA_Col(x, 1.5) //minimum
			col.Resize = SA_ColResizeName(x, col.Name, col.Resize)
			x++
		}
		SA_Col(x, 1) //"+"

		x = 0
		for _, col := range table.Columns {
			if !col.Show {
				continue
			}

			nm := col.Name
			if col.isRowId() {
				nm = "#"
			}

			openDetail := false

			SA_DivStart(x, 0, 1, 1)
			{
				SA_ColMax(0, 100)
				if SA_Button(nm).Show(0, 0, 1, 1).click && !col.isRowId() {
					openDetail = true
				}

				if !col.isRowId() {
					DragAndDropColumn(x, table)
				}
			}
			SA_DivEnd()

			if SA_DialogStart("columnDetail_"+nm, 1, openDetail) {
				ColumnDetail(table, col)
				SA_DialogEnd()
			}

			x++
		}

		//create column
		createColumn := false
		if SA_Button("+").Show(x, 0, 1, 1).click {
			createColumn = true
		}

		if SA_DialogStart("createColumn", 1, createColumn) {

			SA_ColMax(0, 5)
			y := 0
			add_type := ""
			def := ""
			render := ""

			//name
			err := CheckName(store.createColumn, table.FindColumn(store.createColumn) != nil)
			SA_Editbox(&store.createColumn).Error(err).TempToValue(true).ShowDescription(0, y, 1, 1, trns.NAME, 2, 0)
			y++

			//types
			if SA_Button(trns.TEXT).Alpha(1).Align(0).Enable(err == nil).Show(0, y, 1, 1).click { //add icons ...
				add_type = "TEXT"
				def = "''"
			}
			y++

			if SA_Button(trns.FLOAT).Alpha(1).Align(0).Enable(err == nil).Show(0, y, 1, 1).click { //add icons ...
				add_type = "REAL"
				def = "0"
			}
			y++

			//many more column types ...

			if len(add_type) > 0 {
				add_def := "DEFAULT " + def + " NOT NULL"

				SA_SqlWrite("", "ALTER TABLE "+table.Name+" ADD "+store.createColumn+" "+add_type+" "+add_def+";")

				if len(render) > 0 {
					table.Columns = append(table.Columns, &Column{Name: store.createColumn, Type: add_type, Show: true, Resize: 4, Render: render})
					//others will copy 'render' from here
				}

				store.createColumn = ""
				SA_DialogClose()
			}

			SA_DialogEnd()
		}

	}
	SA_DivEnd()

	//rows - extra function ...
	SA_DivStart(0, 1, 1, 1)
	{
		var count int
		{
			query := GetQueryCount(table)
			q := SA_SqlRead("", query)
			q.Next(&count)
		}

		SA_DivSetInfo("scrollOnScreen", 1)
		if table.scrollDown {
			SA_DivSetInfo("scrollVpos", 100000000)
			table.scrollDown = false
		}

		rowSize := table.RowSize + 1

		SA_ColMax(0, 100)
		SA_Row(0, float64(count*rowSize))
		SA_DivStart(0, 0, 1, 1)
		{
			SA_ColMax(0, 100)

			st, en := SA_DivRangeVer(float64(rowSize), -1, -1)

			query, ncols := GetQueryBasic(table)
			var stat *SA_Sql
			if len(query) > 0 {
				stat = SA_SqlRead("", query)
				if stat != nil {
					stat.row_i = uint64(st)
				}
			}
			values := make([]string, ncols)
			args := make([]interface{}, ncols)
			for i := range values {
				args[i] = &values[i]
			}

			for st < en && stat.Next(args...) {

				SA_DivStart(0, st*rowSize, 1, rowSize)
				{
					//columns sizes
					x := 0
					for _, col := range table.Columns {
						if col.Show {
							SA_Col(x, col.Resize)
							x++
						}
					}

					x = 0
					for _, col := range table.Columns {
						if !col.Show {
							continue
						}

						writeCell := false
						if col.isRowId() {

							rowidDialog := SA_Button(values[x]).Show(0, 0, 1, rowSize).click

							if SA_DialogStart("RowId_"+values[x], 1, rowidDialog) {
								SA_ColMax(0, 5)

								//duplicate row ...

								if SA_Button(trns.REMOVE).BackCd(SA_ThemeWarning()).Show(0, 0, 1, 1).click {
									SA_SqlWrite("", "DELETE FROM "+table.Name+" WHERE rowid="+values[x]+";")
									SA_DialogClose()
								}

								SA_DialogEnd()
							}

						} else if col.Type == "BLOB" {

							r, err := strconv.Atoi(values[x])
							if r > 0 && err == nil {

								res := SA_ResourceBuildDbPath("", table.Name, col.Name, r)
								SA_DivStart(x, 0, 1, rowSize)
								{
									SAPaint_File(0, 0, 1, 1, res, "", 0.03, 0, 0, SA_ThemeWhite(), 1, 1, false, false)
									SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeGrey(0.3), 0.03)

									inside := SA_DivInfo("touchInside") > 0
									end := SA_DivInfo("touchEnd") > 0
									if r > 0 && inside {
										SAPaint_Cursor("hand")
									}
									showImageDialog := r > 0 && inside && end

									if SA_DialogStart("Image_"+values[x], 1, showImageDialog) {
										SA_ColMax(0, 15)
										SA_RowMax(0, 15)
										SAPaint_File(0, 0, 1, 1, res, "", 0.03, 0, 0, SA_ThemeWhite(), 1, 1, false, false)
										if SA_DivInfo("touchInside") > 0 && SA_DivInfo("touchEnd") > 0 {
											SA_DialogClose()
										}
										SA_DialogEnd()
									}
								}
								SA_DivEnd()
							}

						} else {
							if SA_Editbox(&values[x]).Margin(0.02).Show(x, 0, 1, rowSize).finished {
								writeCell = true
							}
						}
						//other types ...

						if writeCell {
							v := values[x]
							if col.Type == "TEXT" {
								v = "'" + v + "'"
							}
							SA_SqlWrite("", fmt.Sprintf("UPDATE %s SET %s=%s WHERE rowid=%s;", table.Name, col.Name, v, values[0]))
						}
						x++
					}
				}
				SA_DivEnd()

				st++
			}
		}
		SA_DivEnd()
	}
	SA_DivEnd()

	//add row + column stats - put into extra function ...
	SA_DivStart(0, 2, 1, 1)
	{
		//columns sizes
		{
			x := 0
			for _, col := range table.Columns {
				if col.Show {
					SA_Col(x, col.Resize)
					x++
				}
			}
		}

		var stat *SA_Sql
		q, num_cols := GetQueryStats(table)
		values := make([]string, num_cols)
		if len(q) > 0 {
			stat = SA_SqlRead("", q)

			args := make([]interface{}, num_cols)
			for i := range values {
				args[i] = &values[i]
			}
			stat.Next(args...)
		}

		stat_i := 0
		x := 0
		for _, col := range table.Columns {
			if !col.Show {
				continue
			}

			if col.isRowId() {
				//add row
				if SA_Button("+").Align(1).Title(trns.ADD_ROW).Show(x, 0, 1, 1).click {
					SA_SqlWrite("", "INSERT INTO "+table.Name+" DEFAULT VALUES;")
					table.scrollDown = true
				}
			} else {
				//column stat
				text := ""
				if len(col.StatFunc) > 0 {
					text = col.StatFunc + ": " + values[stat_i]
					stat_i++
				}
				dialogStat := false
				if SA_Button(text).BackCd(SA_ThemeWhite().Aprox(SA_ThemeBack(), 0.4)).Align(0).Show(x, 0, 1, 1).click { //show result
					dialogStat = true
				}
				if SA_DialogStart("Stat_"+strconv.Itoa(x), 1, dialogStat) {

					SA_ColMax(0, 5)
					y := 0
					if col.Type == "INTEGER" || col.Type == "REAL" {

						if SA_Button(trns.MIN).Alpha(1).Align(0).Show(0, y, 1, 1).click {
							col.StatFunc = "min"
							SA_DialogClose()
						}
						y++

						if SA_Button(trns.MAX).Alpha(1).Align(0).Show(0, y, 1, 1).click {
							col.StatFunc = "max"
							SA_DialogClose()
						}
						y++

						if SA_Button(trns.AVG).Alpha(1).Align(0).Show(0, y, 1, 1).click {
							col.StatFunc = "avg"
							SA_DialogClose()
						}
						y++

						if SA_Button(trns.SUM).Alpha(1).Align(0).Show(0, y, 1, 1).click {
							col.StatFunc = "sum"
							SA_DialogClose()
						}
						y++

						if SA_Button(trns.COUNT).Alpha(1).Align(0).Show(0, y, 1, 1).click {
							col.StatFunc = "count"
							SA_DialogClose()
						}
						y++

					}

					SA_DialogEnd()
				}
			}
			x++
		}
	}
	SA_DivEnd()
}

func GetQueryWHERE(table *Table) string {

	var query string

	if table.Filter.Enable {

		nfilters := 0
		for _, f := range table.Filter.Items {
			if f.Column == "" || len(f.GetOpString()) == 0 {
				continue
			}
			nfilters++
		}

		if nfilters > 1 {
			fmt.Print("d")
		}

		i := 0
		queryFilter := ""
		for _, f := range table.Filter.Items {
			op := f.GetOpString()
			if f.Column == "" {
				continue
			}

			val := f.Value
			if len(f.Value) == 0 {
				val = "''"
			}

			queryFilter += f.Column + op + val
			if i+1 < nfilters {
				if table.Filter.Rel == 0 {
					queryFilter += " AND "
				} else {
					queryFilter += " OR "
				}
			}
			i++
		}
		if len(queryFilter) > 0 {
			query += " WHERE " + queryFilter
		}
	}

	if table.Sort.Enable {
		nsorts := 0
		for _, s := range table.Sort.Items {
			if s.Column == "" {
				continue
			}
			nsorts++
		}

		i := 0
		querySort := ""
		for _, s := range table.Sort.Items {
			if s.Column == "" {
				continue
			}

			querySort += s.Column
			if s.Az == 0 {
				querySort += " ASC"
			} else {
				querySort += " DESC"
			}
			if i+1 < nsorts {
				querySort += ","
			}

			i++
		}
		if len(querySort) > 0 {
			query += " ORDER BY " + querySort
		}
	}

	return query
}

func GetQueryCount(table *Table) string {
	query := "SELECT COUNT(*) AS COUNT FROM " + table.Name
	query += GetQueryWHERE(table)
	return query
}

func GetQueryBasic(table *Table) (string, int) {
	query := "SELECT "

	//columns
	ncols := 0
	for _, col := range table.Columns {
		if col.Show {
			ncols++
		}
	}

	if ncols == 0 {
		return "", 0
	}

	i := 0
	for _, col := range table.Columns {
		if !col.Show {
			continue
		}

		switch col.Type {
		case "BLOB":
			query += "rowid AS " + col.Name

		case "DATE":
			query += "DATE(" + col.Name + ")"

		default:
			query += col.Name
		}

		if i+1 < ncols {
			query += ","
		}
		i++
	}

	query += " FROM " + table.Name + ""
	query += GetQueryWHERE(table)
	return query, ncols
}

func GetQueryStats(table *Table) (string, int) {
	query := "SELECT "

	//columns
	ncols := 0
	for _, col := range table.Columns {
		if col.Show && len(col.StatFunc) > 0 {
			ncols++
		}
	}

	if ncols == 0 {
		return "", ncols
	}

	i := 0
	for _, col := range table.Columns {
		if !col.Show || len(col.StatFunc) == 0 {
			continue
		}

		query += col.StatFunc + "(" + col.Name + ")"
		if i+1 < ncols {
			query += ","
		}
		i++
	}

	query += " FROM " + table.Name + ""
	query += GetQueryWHERE(table)
	return query, ncols
}

//export render
func render() uint32 {

	UpdateViews()
	SA_ColMax(0, 100)
	SA_RowMax(1, 100)

	SA_DivStart(0, 0, 1, 1)
	TopHeader()
	SA_DivEnd()

	var selectedTable *Table
	if len(store.Tables) > 0 {
		selectedTable = store.Tables[store.SelectedTable]
	}

	if selectedTable != nil {
		SA_DivStartName(0, 1, 1, 1, selectedTable.Name)
		{
			SA_ColMax(0, 100)
			SA_RowMax(0, 100)

			/*if store.ShowViewPanel {
				SA_Col(1, 1) //minimum
				SA_ColResize(1, 7)
			}*/

			//table
			SA_DivStart(0, 0, 1, 1)
			TableView(selectedTable)
			SA_DivEnd()

			/*if store.ShowViewPanel {
				SA_DivStart(1, 0, 1, 1)
				ViewList(selectedTable)
				SA_DivEnd()
			}*/
		}
		SA_DivEnd()
	}
	return 0
}

func open(buff []byte) bool {
	return false //default json
}
func save() ([]byte, bool) {
	return nil, false //default json
}
func debug() (int, int, string) {
	return -1, 00, "main"
}