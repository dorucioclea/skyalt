package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
)

var store Storage
var trns Translations
var styles SA_Styles

/* -------------------- App information -------------------- */

func SA_InfoFloat(key string) float64 {
	return _sa_info_float(_SA_stringToPtr(key))
}

func SA_InfoSetFloat(key string, v float64) bool {
	return _sa_info_setFloat(_SA_stringToPtr(key), v) > 0
}

func SA_Info(key string) string {
	keyMem := _SA_stringToPtr(key)
	sz := _sa_info_string_len(keyMem)
	if sz > 0 {
		ret := make([]byte, sz)
		if _sa_info_string(keyMem, _SA_bytesToPtr(ret)) > 0 {
			return string(ret)
		}
	}
	return ""
}

func SA_InfoSetVal(key string, value string) int {
	return int(_sa_info_setString(_SA_stringToPtr(key), _SA_stringToPtr(value)))
}
func SA_InfoSet(key string, value string) bool {
	return SA_InfoSetVal(key, value) > 0
}

/* -------------------- Time/Date -------------------- */

func SA_Time() float64 {
	return SA_InfoFloat("time")
}

/* -------------------- Resources -------------------- */

func SA_Resource(path string) []byte {

	pathMem := _SA_stringToPtr(path)

	sz := _sa_resource_len(pathMem)
	if sz > 0 {
		ret := make([]byte, sz)
		if _sa_resource(pathMem, _SA_bytesToPtr(ret)) > 0 {
			return ret
		}
	}
	return nil
}

func SA_ResourceBuildDbPath(db string, table string, column string, row int) string {
	if len(db) == 0 {
		db = SA_Info("db")
	}
	return "db:" + db + "/" + table + "/" + column + "/" + strconv.Itoa(row)
}
func SA_ResourceBuildAssetPath(asset string, file string) string {
	if len(asset) == 0 {
		asset = SA_Info("asset")
	}
	return "asset:" + asset + "/" + file
}

/* -------------------- SQLite storage -------------------- */

type SA_Sql struct {
	db         string
	query      string
	query_hash int64
	row_i      uint64
	row_count  int64
	cache      []byte
}

func SA_SqlRead(db string, query string) *SA_Sql {

	query_hash := _sa_sql_read(_SA_stringToPtr(db), _SA_stringToPtr(query))
	if query_hash == -1 {
		return nil
	}

	var sql SA_Sql
	sql.db = db
	sql.query = query
	sql.query_hash = query_hash
	sql.row_count = _sa_sql_readRowCount(_SA_stringToPtr(sql.db), _SA_stringToPtr(sql.query), sql.query_hash)

	return &sql
}

func (sql *SA_Sql) Next(outs ...interface{}) bool {

	if sql == nil {
		return false
	}

	sz := _sa_sql_readRowLen(_SA_stringToPtr(sql.db), _SA_stringToPtr(sql.query), sql.query_hash, sql.row_i)
	if sz <= 0 {
		return false
	}

	if cap(sql.cache) < int(sz) {
		sql.cache = make([]byte, sz, sz*2)
	} else {
		sql.cache = sql.cache[:sz]
	}

	if _sa_sql_readRow(_SA_stringToPtr(sql.db), _SA_stringToPtr(sql.query), sql.query_hash, sql.row_i, _SA_bytesToPtr(sql.cache)) != 1 {
		return false
	}

	_arrayToArgs(sql.cache, outs...)

	sql.row_i++
	return true
}

func SA_SqlWrite(db string, query string) int64 {
	return _sa_sql_write(_SA_stringToPtr(db), _SA_stringToPtr(query))
}

/* -------------------- Layouts -------------------- */

func SA_ColResize(pos int, val float64) float64 {
	return _sa_div_colResize(uint64(pos), _SA_stringToPtr(""), val)
}
func SA_ColResizeName(pos int, name string, val float64) float64 {
	return _sa_div_colResize(uint64(pos), _SA_stringToPtr(name), val)
}

func SA_RowResize(pos int, val float64) float64 {
	return _sa_div_rowResize(uint64(pos), _SA_stringToPtr(""), val)
}
func SA_RowResizeName(pos int, name string, val float64) float64 {
	return _sa_div_rowResize(uint64(pos), _SA_stringToPtr(name), val)
}

func SA_ColMax(pos int, val float64) float64 {
	return _sa_div_colMax(uint64(pos), val)
}
func SA_RowMax(pos int, val float64) float64 {
	return _sa_div_rowMax(uint64(pos), val)
}
func SA_Col(pos int, val float64) float64 {
	return _sa_div_col(uint64(pos), val)
}
func SA_Row(pos int, val float64) float64 {
	return _sa_div_row(uint64(pos), val)
}

func SA_DivStart(x, y, w, h int) bool {
	return _sa_div_start(uint64(x), uint64(y), uint64(w), uint64(h), _SA_stringToPtr("")) != 0
}
func SA_DivStartName(x, y, w, h int, name string) bool {
	return _sa_div_start(uint64(x), uint64(y), uint64(w), uint64(h), _SA_stringToPtr(name)) != 0
}

func SA_DivEnd() {
	_sa_div_end()
}

func SA_DialogClose() {
	_sa_div_dialogClose()
}

func SA_DialogEnd() {
	_sa_div_dialogEnd()
}

func SA_DialogOpen(name string, tp int) bool {
	return _sa_div_dialogOpen(_SA_stringToPtr(name), uint64(tp)) > 0 //return true if dialog is already opened
}

func SA_DialogStart(name string) bool {
	//maybe create extra api() which will return names of open dialogs ...
	return _sa_div_dialogStart(_SA_stringToPtr(name)) > 0
}

func SA_DivInfoPos(id string, x, y int) float64 {
	return _sa_div_get_info(_SA_stringToPtr(id), int64(x), int64(y))
}
func SA_DivInfo(id string) float64 {
	return SA_DivInfoPos(id, -1, -1)
}

func SA_DivSetInfoPos(id string, val float64, x, y int) float64 {
	return _sa_div_set_info(_SA_stringToPtr(id), val, int64(x), int64(y))
}
func SA_DivSetInfo(id string, val float64) float64 {
	return SA_DivSetInfoPos(id, val, -1, -1)
}

func SA_DivRangeHor(itemSize float64, x, y int) (int, int) {
	wheel := SA_DivInfoPos("layoutStartX", -1, -1)
	screen := SA_DivInfoPos("screenWidth", -1, -1)

	s := wheel / itemSize
	e := (wheel + screen) / itemSize

	return int(s), int(e)
}
func SA_DivRangeVer(itemSize float64, x, y int) (int, int) {
	wheel := SA_DivInfoPos("layoutStartY", -1, -1)
	screen := SA_DivInfoPos("screenHeight", -1, -1)

	s := wheel / itemSize
	e := (wheel + screen) / itemSize

	if e > float64(int(e)) {
		e++
	}
	return int(s), int(e)
}

/* -------------------- Paint -------------------- */

func SAPaint_Rect(x, y, w, h float64, margin float64, cd SACd, borderWidth float64) bool {
	return _sa_paint_rect(x, y, w, h, margin, uint32(cd.R), uint32(cd.G), uint32(cd.G), uint32(cd.A), borderWidth) > 0
}
func SAPaint_Line(sx, sy, ex, ey float64, cd SACd, width float64) bool {
	return _sa_paint_line(0, 0, 1, 1, 0, sx, sy, ex, ey, uint32(cd.R), uint32(cd.G), uint32(cd.G), uint32(cd.A), width) > 0
}
func SAPaint_LineEx(x, y, w, h float64, margin float64, sx, sy, ex, ey float64, cd SACd, width float64) bool {
	return _sa_paint_line(x, y, w, h, margin, sx, sy, ex, ey, uint32(cd.R), uint32(cd.G), uint32(cd.G), uint32(cd.A), width) > 0
}

func SAPaint_Circle(sx, sy, rad float64, cd SACd, borderWidth float64) bool {
	return _sa_paint_circle(0, 0, 1, 1, 0, sx, sy, rad, uint32(cd.R), uint32(cd.G), uint32(cd.G), uint32(cd.A), borderWidth) > 0
}
func SAPaint_CircleEx(x, y, w, h float64, margin float64, sx, sy, rad float64, cd SACd, borderWidth float64) bool {
	return _sa_paint_circle(x, y, w, h, margin, sx, sy, rad, uint32(cd.R), uint32(cd.G), uint32(cd.G), uint32(cd.A), borderWidth) > 0
}

func SAPaint_File(x, y, w, h float64, file string, title string, margin, marginX, marginY float64, cd SACd, alignV, alignH uint32, fill bool, inverse bool) bool {

	var fillB uint32
	var inverseB uint32
	if fill {
		fillB = 1
	}
	if inverse {
		inverseB = 1
	}
	return _sa_paint_file(x, y, w, h, _SA_stringToPtr(file), _SA_stringToPtr(title), margin, marginX, marginY, uint32(cd.R), uint32(cd.G), uint32(cd.G), uint32(cd.A), alignV, alignH, fillB, inverseB) > 0
}

func SAPaint_Text(x, y, w, h float64, value string, margin float64, marginX float64, marginY float64, cd SACd,
	ratioH, lineH float64,
	font, align, alignV uint32,
	selection, edit, tabIsChar, enable bool) bool {
	return _sa_paint_text(x, y, w, h,
		_SA_stringToPtr(value),
		margin, marginX, marginY,
		uint32(cd.R), uint32(cd.G), uint32(cd.G), uint32(cd.A),
		ratioH, lineH, font, align, alignV,
		_SA_boolToUint32(selection), _SA_boolToUint32(edit), _SA_boolToUint32(tabIsChar), _SA_boolToUint32(enable)) > 0
}

func SAPaint_TextWidth(value string, fontId int, ratioH float64, cursorPos int) float64 {
	return _sa_paint_textWidth(_SA_stringToPtr(value), uint32(fontId), ratioH, int64(cursorPos))
}

func SAPaint_TitleEx(x, y, w, h float64, text string) bool {
	return _sa_paint_title(x, y, w, h, _SA_stringToPtr(text)) > 0
}
func SAPaint_Title(text string) bool {
	return SAPaint_TitleEx(0, 0, 1, 1, text)
}

func SAPaint_Cursor(name string) bool {
	return _sa_paint_cursor(_SA_stringToPtr(name)) > 0
}

/* -------------------- Function call -------------------- */

func _argsToArray(data []byte, arg interface{}) []byte {

	switch v := arg.(type) {

	case bool:
		data = append(data, _SA_TpI64)
		if v {
			data = _SA_appendUint64(data, 1)
		} else {
			data = _SA_appendUint64(data, 0)
		}
	case byte:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))
	case int:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))
	case uint:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))

	case int16:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))
	case uint16:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))

	case int32:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))
	case int64:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))

	case uint32:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))
	case uint64:
		data = append(data, _SA_TpI64)
		data = _SA_appendUint64(data, uint64(v))

	case float32:
		data = append(data, _SA_TpF32)
		data = _SA_appendUint64(data, uint64(math.Float32bits(v)))

	case float64:
		data = append(data, _SA_TpF64)
		data = _SA_appendUint64(data, uint64(math.Float64bits(v)))

	case []byte:
		data = append(data, _SA_TpBytes)
		data = _SA_appendUint64(data, uint64(len(v)))
		data = append(data, v...)
	case string:
		data = append(data, _SA_TpBytes)
		data = _SA_appendUint64(data, uint64(len(v)))
		data = append(data, v...)
	}
	return data
}

func _arrayToArgs(args []byte, outs ...interface{}) {
	p := 0
	i := 0
	for p < len(args) && i < len(outs) {

		tp := args[p]
		p += 1

		arg := _SA_getUint64(args[p:])
		p += 8

		switch tp {
		case _SA_TpI32:
			vv := int32(arg)
			switch v := outs[i].(type) {
			case *bool:
				*v = (vv != 0)
			case *int:
				*v = int(vv)
			case *int32:
				*v = int32(vv)
			case *int64:
				*v = int64(vv)
			case *float32:
				*v = float32(vv)
			case *float64:
				*v = float64(vv)
			case *string:
				*v = strconv.Itoa(int(vv))
			}
		case _SA_TpI64:
			vv := int64(arg)
			switch v := outs[i].(type) {
			case *bool:
				*v = (vv != 0)
			case *int:
				*v = int(vv)
			case *int32:
				*v = int32(vv)
			case *int64:
				*v = int64(vv)
			case *float32:
				*v = float32(vv)
			case *float64:
				*v = float64(vv)
			case *string:
				*v = strconv.Itoa(int(vv))
			}
		case _SA_TpF32:
			vv := math.Float32frombits(uint32(arg))
			switch v := outs[i].(type) {
			case *bool:
				*v = (vv != 0)
			case *int:
				*v = int(vv)
			case *int32:
				*v = int32(vv)
			case *int64:
				*v = int64(vv)
			case *float32:
				*v = float32(vv)
			case *float64:
				*v = float64(vv)
			case *string:
				*v = fmt.Sprintf("%f", vv)
			}
		case _SA_TpF64:
			vv := math.Float64frombits(uint64(arg))
			switch v := outs[i].(type) {
			case *bool:
				*v = (vv != 0)
			case *int:
				*v = int(vv)
			case *int32:
				*v = int32(vv)
			case *int64:
				*v = int64(vv)
			case *float32:
				*v = float32(vv)
			case *float64:
				*v = float64(vv)
			case *string:
				*v = fmt.Sprintf("%f", vv)
			}

		case _SA_TpBytes:
			//clone
			arr_n := int(arg)
			arr := make([]byte, arr_n)
			copy(arr, args[p:p+arr_n])
			p += int(arg)

			switch v := outs[i].(type) {
			case *[]byte:
				*v = arr
			case *string:
				//_sa_print(_stringToPtr(string(arr)))
				*v = string(arr)
			}
		}

		i++
	}

	//reset rest
	for i < len(outs) {
		switch v := outs[i].(type) {
		case *bool:
			*v = false
		case *int:
			*v = 0
		case *int32:
			*v = 0
		case *int64:
			*v = 0
		case *float32:
			*v = 0
		case *float64:
			*v = 0
		case *string:
			*v = ""
		}

		i++
	}

}

func SA_CallFn(asset string, fn string, args ...interface{}) int64 {

	//inputs
	data := make([]byte, 0, 256) //pre-alloc
	for _, it := range args {
		data = _argsToArray(data, it)
	}

	//call
	val := _sa_fn_call(_SA_stringToPtr(asset), _SA_stringToPtr(fn), _SA_bytesToPtr(data))

	return val
}

func SA_CallFnShow(x, y, w, h int, asset string, fn string, args ...interface{}) int64 {

	SA_DivStart(x, y, w, h)
	defer SA_DivEnd()

	return SA_CallFn(asset, fn, args...)
}

func SA_CallSetReturn(args ...interface{}) bool {
	data := make([]byte, 0, 256) //pre-alloc
	for _, it := range args {
		data = _argsToArray(data, it)
	}
	return _sa_fn_setReturn(_SA_bytesToPtr(data)) != 0
}

func SA_CallGetReturn(sz int64, outs ...interface{}) bool {
	if sz <= 0 {
		return false
	}
	args := make([]byte, sz)
	_sa_fn_getReturn(_SA_bytesToPtr(args))

	_arrayToArgs(args, outs...)

	return true
}

/* -------------------- Ulits -------------------- */

func SA_Print(str string) {
	_sa_print(_SA_stringToPtr(str))
}
func SA_PrintFloat(val float64) {
	_sa_print_float(val)
}

/* -------------------- SWPs(Skyalt Widgets Proposals) -------------------- */

type _SA_Button struct {
	style *_SA_Style

	value       string
	icon        string
	icon_margin float64
	title       string
	url         string
	enable      bool
}
type _SA_ButtonOut struct {
	click  bool
	rclick bool
}

func SA_ButtonStyle(value string, style *_SA_Style) *_SA_Button {
	var b _SA_Button

	b.value = value
	b.enable = true
	b.style = style

	return &b
}

func SA_Button(value string) *_SA_Button {
	return SA_ButtonStyle(value, &styles.Button)
}
func SA_ButtonLight(value string) *_SA_Button {
	return SA_ButtonStyle(value, &styles.ButtonLight)
}

func SA_ButtonAlpha(value string) *_SA_Button {
	return SA_ButtonStyle(value, &styles.ButtonAlpha)
}
func SA_ButtonMenu(value string) *_SA_Button {
	return SA_ButtonStyle(value, &styles.ButtonMenu)
}

func SA_ButtonAlphaBorder(value string) *_SA_Button {
	return SA_ButtonStyle(value, &styles.ButtonAlphaBorder)
}
func SA_ButtonDanger(value string) *_SA_Button {
	return SA_ButtonStyle(value, &styles.ButtonDanger)
}
func SA_ButtonDangerMenu(value string) *_SA_Button {
	return SA_ButtonStyle(value, &styles.ButtonDangerMenu)
}

func (b *_SA_Button) Value(v string) *_SA_Button {
	b.value = v
	return b
}

func (b *_SA_Button) Highlight(condition bool, style *_SA_Style) *_SA_Button {
	if condition {
		b.style = style
	}
	return b
}

func (b *_SA_Button) Icon(path string, margin float64) *_SA_Button {
	b.icon = path
	b.icon_margin = margin
	return b
}

func (b *_SA_Button) Url(v string) *_SA_Button {
	b.url = v
	return b
}

func (b *_SA_Button) Title(v string) *_SA_Button {
	b.title = v
	return b
}
func (b *_SA_Button) Enable(v bool) *_SA_Button {
	b.enable = v
	return b
}

func (b *_SA_Button) Show(x, y, w, h int) _SA_ButtonOut {

	var ret _SA_ButtonOut

	//SA_DivStart() can trigger sleep mode: no mouse action, outside the screen, etc.
	if SA_DivStart(x, y, w, h) {

		if b.style == nil {
			b.style = &styles.Button //use default
		}

		if b.style.Id == 0 {
			file, err := json.MarshalIndent(b.style, "", "")
			if err != nil {
				return ret
			}
			b.style.Id = uint32(_sa_register_style(_SA_bytesToPtr(file)))
		}

		var out [2 * 8]byte
		_sa_swp_drawButton(b.style.Id, _SA_stringToPtr(b.value), _SA_stringToPtr(b.icon), b.icon_margin, _SA_stringToPtr(b.url), _SA_stringToPtr(b.title), _SA_boolToUint32(b.enable), _SA_bytesToPtr(out[:]))

		ret.click = binary.LittleEndian.Uint64(out[0:]) != 0
		ret.rclick = binary.LittleEndian.Uint64(out[8:]) != 0
	}
	defer SA_DivEnd()

	return ret
}

type _SA_Progress struct {
	value float64
	max   float64

	enable bool
	margin float64

	title string
}

func SA_Progress(value float64) *_SA_Progress {
	var b _SA_Progress

	b.value = value
	b.enable = true
	b.max = 10
	b.margin = 0.03

	return &b
}

func (b *_SA_Progress) Max(v float64) *_SA_Progress {
	b.max = v
	return b
}
func (b *_SA_Progress) Margin(v float64) *_SA_Progress {
	b.margin = v
	return b
}

func (b *_SA_Progress) Show(x, y, w, h int) {

	if SA_DivStart(x, y, w, h) {
		_sa_swp_drawProgress(b.value, b.max, _SA_stringToPtr(b.title), b.margin, _SA_boolToUint32(b.enable))
	}

	defer SA_DivEnd()
}

type _SA_SliderOut struct {
	active   bool
	changed  bool
	finished bool
	size     uint64
}

type _SA_Slider struct {
	value *float64
	min   float64
	max   float64
	jump  float64

	enable bool
	title  string
}

func SA_Slider(value *float64) *_SA_Slider {
	var b _SA_Slider

	b.value = value
	b.enable = true
	b.min = 0
	b.max = 10
	b.jump = 0.1

	return &b
}
func (b *_SA_Slider) Min(v float64) *_SA_Slider {
	b.min = v
	return b
}
func (b *_SA_Slider) Max(v float64) *_SA_Slider {
	b.max = v
	return b
}
func (b *_SA_Slider) Jump(v float64) *_SA_Slider {
	b.jump = v
	return b
}

func (b *_SA_Slider) Show(x, y, w, h int) _SA_SliderOut {

	var ret _SA_SliderOut

	if SA_DivStart(x, y, w, h) {

		var out [3 * 8]byte

		*b.value = _sa_swp_drawSlider(*b.value, b.min, b.max, b.jump, _SA_stringToPtr(b.title), _SA_boolToUint32(b.enable), _SA_bytesToPtr(out[:]))

		ret.active = binary.LittleEndian.Uint64(out[0:]) != 0
		ret.changed = binary.LittleEndian.Uint64(out[8:]) != 0
		ret.finished = binary.LittleEndian.Uint64(out[16:]) != 0

	}

	defer SA_DivEnd()

	return ret
}

type _SA_Text struct {
	value string
	title string

	font    uint32
	frontCd SACd

	margin  float64
	marginX float64
	marginY float64
	align   uint32
	alignV  uint32
	ratioH  float64

	enable    bool
	selection bool

	backCd      SACd
	drawBack    bool
	back_margin float64
}

func SA_Text(value string) *_SA_Text {
	var b _SA_Text

	b.value = value
	b.enable = true
	b.frontCd = SA_ThemeBlack()

	b.margin = 0.03
	b.align = 0
	b.alignV = 1
	b.ratioH = 0.35
	b.selection = true
	b.marginX = 0.1

	return &b
}
func (b *_SA_Text) ValueInt(v int) *_SA_Text {
	b.value = strconv.Itoa(v)
	return b
}

func (b *_SA_Text) ValueFloat(v float64, precision int) *_SA_Text {
	b.value = strconv.FormatFloat(v, 'f', precision, 64)
	return b
}

func (b *_SA_Text) Title(v string) *_SA_Text {
	b.title = v
	return b
}
func (b *_SA_Text) FrontCd(v SACd) *_SA_Text {
	b.frontCd = v
	return b
}
func (b *_SA_Text) BackCd(v SACd, back_margin float64) *_SA_Text {
	b.backCd = v
	b.back_margin = back_margin
	b.drawBack = true
	return b
}

func (b *_SA_Text) Align(v int) *_SA_Text {
	b.align = uint32(v)
	return b
}
func (b *_SA_Text) AlignV(v int) *_SA_Text {
	b.alignV = uint32(v)
	return b
}

func (b *_SA_Text) RatioH(v float64) *_SA_Text {
	b.ratioH = v
	return b
}

func (b *_SA_Text) ShowDescription(x, y, w, h int, description string, width float64, align int) {

	if SA_DivStart(x, y, w, h) {
		if width > 0 {
			//1 row
			SA_Col(0, width)
			SA_ColMax(1, 100)
			SA_Text(description).Align(align).Show(0, 0, 1, 1)
			b.Show(1, 0, 1, 1)
		} else {
			//2 rows
			SA_ColMax(0, 100)
			SA_Text(description).Align(align).Show(0, 0, 1, 1)
			b.Show(0, 1, 1, 1)
		}
	}
	SA_DivEnd()
}

func (b *_SA_Text) Show(x, y, w, h int) {
	if SA_DivStart(x, y, w, h) {

		if b.drawBack {
			SAPaint_Rect(0, 0, 1, 1, b.back_margin, b.backCd, 0)
		}

		_sa_swp_drawText(uint32(b.frontCd.R), uint32(b.frontCd.G), uint32(b.frontCd.B), uint32(b.frontCd.A),
			_SA_stringToPtr(b.value), _SA_stringToPtr(b.title), b.font,
			b.margin, b.marginX, b.marginY, b.align, b.alignV, b.ratioH,
			_SA_boolToUint32(b.enable), _SA_boolToUint32(b.selection))
	}
	defer SA_DivEnd()
}
func (b *_SA_Text) DrawPaint(x, y, w, h float64) {
	SAPaint_Text(x, y, w, h, b.value, b.margin, b.marginX, b.marginY, b.frontCd, b.ratioH, 1, b.font, b.align, b.alignV, b.selection, false, false, true)
}

type _SA_Editbox struct {
	value     interface{}
	valueOrig string
	title     string

	valueOrigSet bool

	font uint32

	backCd  SACd
	frontCd SACd

	margin  float64
	marginX float64
	marginY float64
	align   uint32
	alignV  uint32
	ratioH  float64

	enable      bool
	tempToValue bool
	asNumber    bool

	drawBack   bool
	drawBorder bool
	ghost      string
	precision  int

	err error
}
type _SA_EditboxOut struct {
	active   bool
	changed  bool
	finished bool
	size     uint64
}

func SA_Editbox(value interface{}) *_SA_Editbox {
	var b _SA_Editbox

	b.value = value

	b.enable = true
	b.backCd = SA_ThemeWhite()
	b.frontCd = SA_ThemeBlack()

	b.margin = 0.03
	b.align = 0
	b.alignV = 1
	b.ratioH = 0.35
	b.marginX = 0.1
	b.drawBack = true
	b.drawBorder = true
	b.precision = 3

	return &b
}

func (b *_SA_Editbox) ValueOrig(v string) *_SA_Editbox {
	b.valueOrig = v
	b.valueOrigSet = true
	return b
}

func (b *_SA_Editbox) Align(v uint32) *_SA_Editbox {
	b.align = v
	return b
}
func (b *_SA_Editbox) Margin(v float64) *_SA_Editbox {
	b.margin = v
	return b
}

func (b *_SA_Editbox) TempToValue(v bool) *_SA_Editbox {
	b.tempToValue = v
	return b
}
func (b *_SA_Editbox) AsNumber(v bool) *_SA_Editbox {
	b.asNumber = v
	return b
}
func (b *_SA_Editbox) Precision(v int) *_SA_Editbox {
	b.precision = v
	return b
}

func (b *_SA_Editbox) Ghost(v string) *_SA_Editbox {
	b.ghost = v
	return b
}

func (b *_SA_Editbox) Error(v error) *_SA_Editbox {
	b.err = v
	return b
}

func (b *_SA_Editbox) Enable(v bool) *_SA_Editbox {
	b.enable = v
	return b
}

func (b *_SA_Editbox) DrawBorder(v bool) *_SA_Editbox {
	b.drawBorder = v
	return b
}

func (b *_SA_Editbox) DrawBack(v bool) *_SA_Editbox {
	b.drawBack = v
	return b
}
func (b *_SA_Editbox) BackCd(v SACd) *_SA_Editbox {
	b.backCd = v
	return b
}

func (b *_SA_Editbox) Highlight(cd SACd) *_SA_Editbox {
	b.DrawBack(true)
	b.BackCd(cd)
	return b
}

func (b *_SA_Editbox) HighlightError(enable bool) *_SA_Editbox {
	if enable {
		b.DrawBack(true)
		b.BackCd(SA_ThemeError())
	}
	return b
}
func (b *_SA_Editbox) HighlightWarning(enable bool) *_SA_Editbox {
	if enable {
		b.DrawBack(true)
		b.BackCd(SA_ThemeWarning())
	}
	return b
}
func (b *_SA_Editbox) HighlightEdit(enable bool) *_SA_Editbox {
	if enable {
		b.DrawBack(true)
		b.BackCd(SA_ThemeEdit())
	}
	return b
}

func (b *_SA_Editbox) ShowDescription(x, y, w, h int, description string, width float64, align int) _SA_EditboxOut {

	var ret _SA_EditboxOut
	if SA_DivStart(x, y, w, h) {
		if width > 0 {
			//1 row
			SA_Col(0, width)
			SA_ColMax(1, 100)
			SA_Text(description).Align(align).Show(0, 0, 1, 1)
			ret = b.Show(1, 0, 1, 1)
		} else {
			//2 rows
			SA_ColMax(0, 100)
			SA_Text(description).Align(align).Show(0, 0, 1, 1)
			ret = b.Show(0, 1, 1, 1)
		}
	}
	SA_DivEnd()

	return ret
}

func (b *_SA_Editbox) Show(x, y, w, h int) _SA_EditboxOut {

	var ret _SA_EditboxOut

	if SA_DivStart(x, y, w, h) {
		if b.drawBack {
			SAPaint_Rect(0, 0, 1, 1, b.margin, b.backCd, 0)
		}

		if b.err != nil {
			SAPaint_Rect(0, 0, 1, 1, b.margin, SA_ThemeError(), 0)
		}

		value := ""
		switch v := b.value.(type) {
		case *float64:
			value = strconv.FormatFloat(*v, 'f', b.precision, 64)
		case *int:
			value = strconv.Itoa(*v)
		case *string:
			value = *v
			//float32, byte, etc ...
		}

		valueOrig := value
		if b.valueOrigSet {
			valueOrig = b.valueOrig
		}

		title := ""
		if b.err != nil {
			title = b.err.Error()
		} else if len(b.title) > 0 {
			title = b.title
		}

		var out [4 * 8]byte
		_sa_swp_drawEdit(uint32(b.frontCd.R), uint32(b.frontCd.G), uint32(b.frontCd.B), uint32(b.frontCd.A),
			_SA_stringToPtr(value), _SA_stringToPtr(valueOrig), _SA_stringToPtr(title), b.font,
			b.margin, b.marginX, b.marginY, b.align, b.alignV, b.ratioH,
			_SA_boolToUint32(b.enable),
			_SA_bytesToPtr(out[:]))

		ret.active = binary.LittleEndian.Uint64(out[0:]) != 0
		ret.changed = binary.LittleEndian.Uint64(out[8:]) != 0
		ret.finished = binary.LittleEndian.Uint64(out[16:]) != 0
		ret.size = binary.LittleEndian.Uint64(out[24:])

		if ret.finished || (b.tempToValue && ret.changed) {
			val := make([]byte, ret.size)
			_sa_swp_getEditValue(_SA_bytesToPtr(val))

			switch v := b.value.(type) {
			case *float64:
				*v, _ = strconv.ParseFloat(string(val), 64)
			case *int:
				*v, _ = strconv.Atoi(string(val))
			case *string:
				*v = string(val)
				//float32, byte, etc ...
			}
		}

		//ghost
		if len(b.ghost) > 0 && ret.size == 0 {
			SAPaint_Text(0, 0, 1, 1, b.ghost, b.margin, b.marginX, b.marginY, b.backCd.Aprox(b.frontCd, 0.5), b.ratioH, 1, b.font, 1, 1, false, false, false, b.enable)
		}

		if b.drawBorder {
			cd := b.frontCd
			if !b.enable {
				cd = SA_ThemeWhite().Aprox(cd, 0.3)
			}
			SAPaint_Rect(0, 0, 1, 1, b.margin, b.backCd.Aprox(cd, 0.7), 0.03)
		}

	}
	defer SA_DivEnd()

	return ret
}

type _SA_Combo struct {
	value   *int
	options string

	title string

	font uint32

	cd      SACd
	frontCd SACd

	margin  float64
	marginX float64
	marginY float64
	align   uint32
	ratioH  float64

	search bool //...
	err    error

	enable bool
}

func SA_Combo(value *int, options string) *_SA_Combo {
	var b _SA_Combo

	b.value = value
	b.options = options

	b.enable = true
	b.frontCd = SA_ThemeBlack()

	b.margin = 0.03
	b.align = 0
	b.ratioH = 0.35
	b.marginX = 0.1
	return &b
}
func (b *_SA_Combo) Enable(v bool) *_SA_Combo {
	b.enable = v
	return b
}
func (b *_SA_Combo) Search(v bool) *_SA_Combo {
	b.search = v
	return b
}
func (b *_SA_Combo) Align(v uint32) *_SA_Combo {
	b.align = v
	return b
}

func (b *_SA_Combo) Error(v error) *_SA_Combo {
	b.err = v
	return b
}

func (b *_SA_Combo) ShowDescription(x, y, w, h int, description string, width float64, align int) bool {

	var ret bool
	if SA_DivStart(x, y, w, h) {
		if width > 0 {
			//1 row
			SA_Col(0, width)
			SA_ColMax(1, 100)
			SA_Text(description).Align(align).Show(0, 0, 1, 1)
			ret = b.Show(1, 0, 1, 1)
		} else {
			//2 rows
			SA_ColMax(0, 100)
			SA_Text(description).Align(align).Show(0, 0, 1, 1)
			ret = b.Show(0, 1, 1, 1)
		}
	}
	SA_DivEnd()

	return ret
}

func (b *_SA_Combo) Show(x, y, w, h int) bool {

	changed := false

	if SA_DivStart(x, y, w, h) {

		if b.err != nil {
			SAPaint_Rect(0, 0, 1, 1, 0, SA_ThemeError(), 0)
		}

		title := ""
		if b.err != nil {
			title = b.err.Error()
		} else if len(b.title) > 0 {
			title = b.title
		}

		v := _sa_swp_drawCombo(uint32(b.frontCd.R), uint32(b.frontCd.G), uint32(b.frontCd.B), uint32(b.frontCd.A),
			uint64(*b.value), _SA_stringToPtr(b.options), _SA_stringToPtr(title), b.font,
			b.margin, b.marginX, b.marginY, b.align, b.ratioH,
			_SA_boolToUint32(b.enable))

		changed = *b.value != int(v)
		*b.value = int(v)
	}
	SA_DivEnd()

	return changed
}

type _SA_Checkbox struct {
	value       *bool
	enable      bool
	description string

	backCd  SACd
	frontCd SACd

	height float64
	align  uint32
	alignV uint32

	title string
}

func SA_Checkbox(value *bool, description string) *_SA_Checkbox {
	var b _SA_Checkbox

	b.value = value
	b.description = description

	b.height = 1
	b.alignV = 1
	b.enable = true
	b.frontCd = SA_ThemeBlack()

	return &b
}

func (b *_SA_Checkbox) Align(v int) *_SA_Checkbox {
	b.align = uint32(v)
	return b
}

func (b *_SA_Checkbox) Show(x, y, w, h int) bool {

	changed := false

	if SA_DivStart(x, y, w, h) {

		val := uint64(0)
		if *b.value {
			val = 1
		}

		v := _sa_swp_drawCheckbox(uint32(b.frontCd.R), uint32(b.frontCd.G), uint32(b.frontCd.B), uint32(b.frontCd.A),
			val, _SA_stringToPtr(b.description), _SA_stringToPtr(b.title), b.height, b.align, b.alignV, _SA_boolToUint32(b.enable))

		changed = (val != uint64(v))
		if changed {
			*b.value = !(*b.value)
		}
	}
	defer SA_DivEnd()

	return changed
}

type _SA_Image struct {
	file  string
	title string

	margin  float64
	marginX float64
	marginY float64
	align   uint32
	alignV  uint32

	enable    bool
	fill      bool
	inverseCd bool
	cd        SACd
}

func SA_Image(file string) *_SA_Image {
	var b _SA_Image

	b.file = file
	b.enable = true
	b.cd = SA_ThemeBlack()

	b.margin = 0.03
	b.align = 1
	b.alignV = 1
	b.marginX = 0.1

	return &b
}

func (b *_SA_Image) AlignV(v int) *_SA_Image {
	b.alignV = uint32(v)
	return b
}

func (b *_SA_Image) Margin(v float64) *_SA_Image {
	b.margin = v
	return b
}

func (b *_SA_Image) InverseColor(v bool) *_SA_Image {
	b.inverseCd = v
	return b
}

func (b *_SA_Image) Show(x, y, w, h int) {

	if SA_DivStart(x, y, w, h) {
		_sa_paint_file(0, 0, 1, 1,
			_SA_stringToPtr(b.file), _SA_stringToPtr(b.title), b.margin, b.marginX, b.marginY,
			uint32(b.cd.R), uint32(b.cd.G), uint32(b.cd.B), uint32(b.cd.A),
			b.align, b.alignV, _SA_boolToUint32(b.fill), _SA_boolToUint32(b.inverseCd))

	}
	defer SA_DivEnd()
}

/* -------------------- Themes, Colors, etc. -------------------- */
type SACd struct {
	R, G, B, A byte
}

func SA_InitCd(r uint32, g uint32, b uint32, a uint32) SACd {
	return SACd{byte(r), byte(g), byte(b), byte(a)}
}
func (s SACd) Aprox(e SACd, t float32) SACd {
	var ret SACd
	ret.R = byte(float32(s.R) + (float32(e.R)-float32(s.R))*t)
	ret.G = byte(float32(s.G) + (float32(e.G)-float32(s.G))*t)
	ret.B = byte(float32(s.B) + (float32(e.B)-float32(s.B))*t)
	ret.A = byte(float32(s.A) + (float32(e.A)-float32(s.A))*t)
	return ret
}

func SA_ThemeCd() SACd {

	cd := SACd{90, 180, 180, 255} // ocean
	switch SA_InfoFloat("theme") {
	case 1:
		cd = SACd{200, 100, 80, 255}
	case 2:
		cd = SACd{130, 170, 210, 255}
	case 3:
		cd = SACd{130, 180, 130, 255}
	case 4:
		cd = SACd{160, 160, 160, 255}
	}
	return cd
}

func SA_ThemeBack() SACd {
	return SACd{220, 220, 220, 255}
}
func SA_ThemeWhite() SACd {
	return SACd{255, 255, 255, 255}
}
func SA_ThemeMid() SACd {
	return SACd{127, 127, 127, 255}
}
func SA_ThemeBlack() SACd {
	return SACd{0, 0, 0, 255}
}
func SA_ThemeGrey(t float64) SACd {
	return SACd{byte(255 * t), byte(255 * t), byte(255 * t), 255}
}
func SA_ThemeEdit() SACd {
	return SACd{210, 110, 90, 255}
}
func SA_ThemeWarning() SACd {
	return SACd{230, 110, 50, 255}
}

func SA_ThemeError() SACd {
	return SACd{230, 70, 70, 255}
}

/* -------------------- Helpers :) -------------------- */

const _SA_TpI32 = byte(0x7f)
const _SA_TpI64 = byte(0x7e)
const _SA_TpF32 = byte(0x7d)
const _SA_TpF64 = byte(0x7c)
const _SA_TpBytes = byte(0x7b)

func _SA_putUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[0] = byte(v)
	b[1] = byte(v >> 8)
	b[2] = byte(v >> 16)
	b[3] = byte(v >> 24)
	b[4] = byte(v >> 32)
	b[5] = byte(v >> 40)
	b[6] = byte(v >> 48)
	b[7] = byte(v >> 56)
}

func _SA_appendUint64(b []byte, v uint64) []byte {
	return append(b,
		byte(v),
		byte(v>>8),
		byte(v>>16),
		byte(v>>24),
		byte(v>>32),
		byte(v>>40),
		byte(v>>48),
		byte(v>>56),
	)
}

func _SA_getUint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[0]) | uint64(b[1])<<8 | uint64(b[2])<<16 | uint64(b[3])<<24 |
		uint64(b[4])<<32 | uint64(b[5])<<40 | uint64(b[6])<<48 | uint64(b[7])<<56
}

func _SA_boolToUint32(v bool) uint32 {
	if v {
		return 1
	}
	return 0
}

func SA_Var(value interface{}, buff *[]byte, w bool) bool {

	if w {
		//write
		switch v := value.(type) {
		case *bool:
			if *v {
				*buff = _SA_appendUint64(*buff, 1)
			} else {
				*buff = _SA_appendUint64(*buff, 0)

			}
		case *byte:
			*buff = _SA_appendUint64(*buff, uint64(*v))
		case *int:
			*buff = _SA_appendUint64(*buff, uint64(*v))
		case *int8:
			*buff = _SA_appendUint64(*buff, uint64(*v))
		case *int16:
			*buff = _SA_appendUint64(*buff, uint64(*v))
		case *int32:
			*buff = _SA_appendUint64(*buff, uint64(*v))
		case *int64:
			*buff = _SA_appendUint64(*buff, uint64(*v))
		case *float32:
			*buff = _SA_appendUint64(*buff, uint64(math.Float32bits(*v)))
		case *float64:
			*buff = _SA_appendUint64(*buff, uint64(math.Float64bits(*v)))

		case *[]byte:
			*buff = _SA_appendUint64(*buff, uint64(len(*v)))
			*buff = append(*buff, (*v)...)

		case *string:
			*buff = _SA_appendUint64(*buff, uint64(len(*v)))
			*buff = append(*buff, (*v)...)
		}
	} else {
		if len(*buff) < 8 {
			return false
		}
		arg := _SA_getUint64(*buff)
		*buff = (*buff)[8:]

		switch v := value.(type) {
		case *bool:
			if arg != 0 {
				*v = true
			} else {
				*v = false
			}
		case *byte:
			*v = byte(arg)
		case *int:
			*v = int(arg)
		case *int8:
			*v = int8(arg)
		case *int16:
			*v = int16(arg)
		case *int32:
			*v = int32(arg)
		case *int64:
			*v = int64(arg)
		case *float32:
			*v = math.Float32frombits(uint32(arg))
		case *float64:
			*v = math.Float64frombits(uint64(arg))

		case *[]byte:
			*v = make([]byte, arg)
			if len(*buff) < int(arg) {
				return false
			}
			copy(*v, (*buff)[:arg])
			*buff = (*buff)[arg:]

		case *string:
			vb := make([]byte, arg)
			if len(*buff) < int(arg) {
				return false
			}
			copy(vb, (*buff)[:arg])
			*buff = (*buff)[arg:]
			*v = string(vb)
		}
	}

	return true
}

func SA_RowSpacer(x, y, w, h int) {
	//SA_Row(y, row)

	SA_DivStart(x, y, w, h)
	grey := byte(180)
	SAPaint_Line(0, 0.5, 1, 0.5, SACd{grey, grey, grey, 255}, 0.03)
	SA_DivEnd()
}

func SA_ColSpacer(x, y, w, h int) {
	//SA_Col(x, col)

	SA_DivStart(x, y, w, h)
	grey := byte(180)
	SAPaint_Line(0.5, 0, 0.5, 1, SACd{grey, grey, grey, 255}, 0.03)
	SA_DivEnd()
}

func SA_DialogConfirm() bool {
	SA_ColMax(0, 5)

	click := SA_ButtonDanger("Confirm").Show(0, 0, 1, 1).click //translations ... maybe add 'confirm string' do args ...
	if click {
		SA_DialogClose()
	}
	return click
}

type SA_Drop_POS int

const (
	SA_Drop_INSIDE  SA_Drop_POS = 0
	SA_Drop_V_LEFT  SA_Drop_POS = 1
	SA_Drop_V_RIGHT SA_Drop_POS = 2
	SA_Drop_H_LEFT  SA_Drop_POS = 3
	SA_Drop_H_RIGHT SA_Drop_POS = 4
)

func SA_Div_SetDrag(group string, id uint64) bool {
	return _sa_div_drag(_SA_stringToPtr(group), id) > 0
}

func SA_Div_IsDrop(group string, vertical, horizontal, inside bool) (uint64, SA_Drop_POS, bool) {
	var out [2 * 8]byte

	done := _sa_div_drop(_SA_stringToPtr(group), _SA_boolToUint32(vertical), _SA_boolToUint32(horizontal), _SA_boolToUint32(inside), _SA_bytesToPtr(out[:]))

	id := binary.LittleEndian.Uint64(out[0:])
	pos := SA_Drop_POS(binary.LittleEndian.Uint64(out[8:]))
	return id, pos, done > 0
}

// usefull for moving element inside array for Drag & Drop
func SA_MoveElement[T any](array_src *[]T, array_dst *[]T, src int, dst int, pos SA_Drop_POS) {

	//check
	if src < dst && (pos == SA_Drop_V_LEFT || pos == SA_Drop_H_LEFT) {
		dst--
	}
	if src > dst && (pos == SA_Drop_V_RIGHT || pos == SA_Drop_H_RIGHT) {
		dst++
	}

	//move(by swap one-by-one)
	if array_src == array_dst {
		for i := src; i < dst; i++ {
			(*array_dst)[i], (*array_dst)[i+1] = (*array_dst)[i+1], (*array_dst)[i]
		}
		for i := src; i > dst; i-- {
			(*array_dst)[i], (*array_dst)[i-1] = (*array_dst)[i-1], (*array_dst)[i]
		}
	} else {

		backup := (*array_src)[src]

		//remove
		*array_src = append((*array_src)[:src], (*array_src)[src+1:]...)

		//insert
		if dst < len(*array_dst) {
			*array_dst = append((*array_dst)[:dst+1], (*array_dst)[dst:]...)
			(*array_dst)[dst] = backup
		} else {
			*array_dst = append(*array_dst, backup)
		}
	}
}

func SA_RenderApp(app string, db string, sts_id int) bool {
	return _sa_render_app(_SA_stringToPtr(app), _SA_stringToPtr(db), uint64(sts_id)) >= 0
}

func SA_Rating(value int, max_value int, cdActive SACd, cdDeactive SACd, icon string) (int, bool) {

	changed := false

	SA_DivSetInfo("scrollHnarrow", 1)
	SA_DivSetInfo("scrollVshow", 0)

	w := SA_DivInfo("layoutWidth") / float64(max_value)
	h := SA_DivInfo("layoutHeight")

	if w < 0.7 {
		w = 0.7
	}

	SA_Row(0, h)
	for i := 0; i < max_value; i++ {
		SA_Col(i, w) //1
	}

	for i := 0; i < max_value; i++ {
		SA_DivStart(i, 0, 1, 1)
		{
			cd := cdActive
			if i >= value {
				cd = cdDeactive
			}

			active := SA_DivInfo("touchActive") > 0
			inside := SA_DivInfo("touchInside") > 0
			end := SA_DivInfo("touchEnd") > 0
			touch_x := SA_DivInfo("touchX")

			if active || inside {
				cd = cd.Aprox(cdActive, 0.4)
				SAPaint_Cursor("hand")
			}

			if active && inside {
				cd = cd.Aprox(cdActive, 0.7)
			}

			if inside && end {
				old_value := value
				if i == 0 && touch_x < 0.25 {
					value = 0
				} else {
					value = i + 1
				}
				changed = (value != old_value)
			}

			SAPaint_File(0, 0, 1, 1, icon, "", 0.1, 0, 0, cd, 1, 1, false, false)
		}
		SA_DivEnd()
	}

	return value, changed
}

/* -------------------- Styles -------------------- */

type _SAStyle_Div struct {
	Margin_top, Margin_bottom, Margin_left, Margin_right     float64 //from cell
	Border_top, Border_bottom, Border_left, Border_right     float64 //from cell
	Padding_top, Padding_bottom, Padding_left, Padding_right float64 //from cell

	Margin_top_color, Margin_bottom_color, Margin_left_color, Margin_right_color     SACd
	Border_top_color, Border_bottom_color, Border_left_color, Border_right_color     SACd
	Padding_top_color, Padding_bottom_color, Padding_left_color, Padding_right_color SACd
	Content_color                                                                    SACd

	Image_color                SACd
	Image_margin               float64
	Image_fill                 bool
	Image_alignV, Image_alignH int

	Font_color               SACd
	Font_path                string
	Font_height              float64 //from cell
	Font_alignV, Font_alignH int

	Cursor string

	//radius ...
	//shadow ...
	//transition_sec(blend between states) ...
}

func (b *_SAStyle_Div) MarginEx(top, bottom, left, right float64) *_SAStyle_Div {
	b.Margin_top = top
	b.Margin_bottom = bottom
	b.Margin_left = left
	b.Margin_right = right
	return b
}
func (b *_SAStyle_Div) Margin(v float64) *_SAStyle_Div {
	return b.MarginEx(v, v, v, v)
}

func (b *_SAStyle_Div) BorderEx(top, bottom, left, right float64) *_SAStyle_Div {
	b.Border_top = top
	b.Border_bottom = bottom
	b.Border_left = left
	b.Border_right = right
	return b
}
func (b *_SAStyle_Div) Border(v float64) *_SAStyle_Div {
	return b.BorderEx(v, v, v, v)
}

func (b *_SAStyle_Div) PaddingEx(top, bottom, left, right float64) *_SAStyle_Div {
	b.Padding_top = top
	b.Padding_bottom = bottom
	b.Padding_left = left
	b.Padding_right = right
	return b
}
func (b *_SAStyle_Div) Padding(v float64) *_SAStyle_Div {
	return b.PaddingEx(v, v, v, v)
}

func (b *_SAStyle_Div) MarginColorEx(top, bottom, left, right SACd) *_SAStyle_Div {
	b.Margin_top_color = top
	b.Margin_bottom_color = bottom
	b.Margin_left_color = left
	b.Margin_right_color = right
	return b
}
func (b *_SAStyle_Div) MarginColor(v SACd) *_SAStyle_Div {
	return b.MarginColorEx(v, v, v, v)
}

func (b *_SAStyle_Div) BorderColorEx(top, bottom, left, right SACd) *_SAStyle_Div {
	b.Border_top_color = top
	b.Border_bottom_color = bottom
	b.Border_left_color = left
	b.Border_right_color = right
	return b
}
func (b *_SAStyle_Div) BorderColor(v SACd) *_SAStyle_Div {
	return b.BorderColorEx(v, v, v, v)
}

func (b *_SAStyle_Div) PaddingColorEx(top, bottom, left, right SACd) *_SAStyle_Div {
	b.Padding_top_color = top
	b.Padding_bottom_color = bottom
	b.Padding_left_color = left
	b.Padding_right_color = right
	return b
}
func (b *_SAStyle_Div) PaddingColor(v SACd) *_SAStyle_Div {
	return b.PaddingColorEx(v, v, v, v)
}

type _SA_Style struct {
	Id           uint32
	Main         _SAStyle_Div
	Hover        _SAStyle_Div
	Touch_hover  _SAStyle_Div
	Touch_ouside _SAStyle_Div
	Disable      _SAStyle_Div
}

func (b *_SA_Style) Padding(v float64) *_SA_Style {
	b.Main.Padding(v)
	b.Hover.Padding(v)
	b.Touch_hover.Padding(v)
	b.Touch_ouside.Padding(v)
	b.Disable.Padding(v)
	return b
}
func (b *_SA_Style) Margin(v float64) *_SA_Style {
	b.Main.Margin(v)
	b.Hover.Margin(v)
	b.Touch_hover.Margin(v)
	b.Touch_ouside.Margin(v)
	b.Disable.Margin(v)
	return b
}

func (b *_SA_Style) FontAlignH(v int) *_SA_Style {
	b.Main.Font_alignH = v
	b.Hover.Font_alignH = v
	b.Touch_hover.Font_alignH = v
	b.Touch_ouside.Font_alignH = v
	b.Disable.Font_alignH = v
	return b
}
func (b *_SA_Style) FontAlignV(v int) *_SA_Style {
	b.Main.Font_alignV = v
	b.Hover.Font_alignV = v
	b.Touch_hover.Font_alignV = v
	b.Touch_ouside.Font_alignV = v
	b.Disable.Font_alignV = v
	return b
}

func (b *_SA_Style) FontH(v float64) *_SA_Style {
	b.Main.Font_height = v
	b.Hover.Font_height = v
	b.Touch_hover.Font_height = v
	b.Touch_ouside.Font_height = v
	b.Disable.Font_height = v
	return b
}

//more ...

type SA_Styles struct {
	Button             _SA_Style
	ButtonLight        _SA_Style
	ButtonAlpha        _SA_Style
	ButtonMenu         _SA_Style
	ButtonMenuSelected _SA_Style

	ButtonBig      _SA_Style
	ButtonLightBig _SA_Style
	ButtonAlphaBig _SA_Style
	ButtonMenuBig  _SA_Style

	ButtonAlphaBorder _SA_Style
	ButtonDanger      _SA_Style
	ButtonDangerMenu  _SA_Style

	ButtonIcon _SA_Style
}
