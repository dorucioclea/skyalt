package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net"
	"reflect"
	"strconv"
)

var conn *net.TCPConn

func main() {
	port, sts_id, asset := debug()

	if port < 0 {
		port = 8091
	}

	tcpServer, err := net.ResolveTCPAddr("tcp", "localhost:"+strconv.Itoa(port))
	if err != nil {
		fmt.Printf("ResolveTCPAddr() failed: %v\n", err)
		return
	}

	conn, err = net.DialTCP("tcp", nil, tcpServer)
	if err != nil {
		fmt.Printf("DialTCP() failed: %v\n", err)
		return
	}

	WriteUint64(uint64(sts_id))
	WriteBytes([]byte(asset))

	fmt.Printf("Connected on port: %d\n", port)

	for {
		fnName, err := ReadBytes()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Panic(err)
		}
		args, err := ReadBytes()
		if err != nil {
			log.Panic(err)
		}

		switch string(fnName) {
		case "render":
			render()

		case "_sa_open":
			var js []byte
			_arrayToArgs(args, &js)
			if !open(js) {
				json.Unmarshal(js, &store)
			}

		case "_sa_save":
			js, written := save()
			if !written {
				js, _ = json.MarshalIndent(&store, "", "")
			}
			_sa_storage_write(_SA_bytesToPtr(js))

		case "_sa_translations_set":
			e := reflect.ValueOf(&trns).Elem()
			for i := 0; i < e.NumField(); i++ {
				e.Field(i).SetString("{" + e.Type().Field(i).Name + "}")
			}

			var js []byte
			_arrayToArgs(args, &js)
			json.Unmarshal(js, &trns)

		default:
			log.Panic("Unknown function: ", string(fnName))
		}

		WriteUint64(100) //end of function call
	}

	err = conn.Close()
	if err != nil {
		fmt.Printf("Close() failed: %v\n", err)
		return
	}
}

func WriteUint64(v uint64) {
	var b [8]byte
	_SA_putUint64(b[:], v)

	_, err := conn.Write(b[:])
	if err != nil {
		log.Panic(err)
	}
}

func ReadUint64e() (uint64, error) {
	var b [8]byte
	_, err := conn.Read(b[:])
	if err != nil {
		return 0, err
	}
	return _SA_getUint64(b[:]), nil
}

func ReadUint64() uint64 {
	v, err := ReadUint64e()
	if err != nil {
		log.Panic(err)
	}
	return v
}

func WriteFloat64(v float64) {
	WriteUint64(math.Float64bits(v))
}

func ReadFloat64() float64 {
	return math.Float64frombits(ReadUint64())
}

func WriteMem(mem SAMem) {
	//data := _SA_ptrToBytes(mem)

	WriteUint64(uint64(len(mem.v))) //size
	_, err := conn.Write(mem.v)     //data
	if err != nil {
		log.Panic(err)
	}
}

func ReadBytes() ([]byte, error) {
	sz, err := ReadUint64e()
	if err != nil {
		return nil, err
	}

	data := make([]byte, sz)
	_, err = conn.Read(data)
	if err != nil {
		return nil, err
	}
	return data, nil
}
func WriteBytes(data []byte) {
	WriteUint64(uint64(len(data))) //size
	_, err := conn.Write(data)     //data
	if err != nil {
		log.Panic(err)
	}
}

func ReadMem(mem SAMem) {
	sz := int(ReadUint64())
	if sz != len(mem.v) {
		log.Panic("Wrong size")
	}
	_, err := conn.Read(mem.v)
	if err != nil {
		log.Panic(err)
	}
}

//-------

func _sa_storage_write(jsonMem SAMem) int64 {
	WriteUint64(0)
	WriteMem(jsonMem)
	return int64(ReadUint64())
}

func _sa_info_float(keyMem SAMem) float64 {
	WriteUint64(1)
	WriteMem(keyMem)
	return ReadFloat64()
}

func _sa_info_setFloat(keyMem SAMem, value float64) int64 {
	WriteUint64(2)
	WriteMem(keyMem)
	WriteFloat64(value)
	return int64(ReadUint64())
}

func _sa_info_string(keyMem SAMem, dstMem SAMem) int64 {
	WriteUint64(3)
	WriteMem(keyMem)

	ReadMem(dstMem)
	return int64(ReadUint64())
}

func _sa_info_string_len(keyMem SAMem) int64 {
	WriteUint64(4)
	WriteMem(keyMem)
	return int64(ReadUint64())
}

func _sa_info_setString(keyMem SAMem, valueMem SAMem) int64 {
	WriteUint64(5)
	WriteMem(keyMem)
	WriteMem(valueMem)
	return int64(ReadUint64())
}

func _sa_resource(pathMem SAMem, dstMem SAMem) int64 {
	WriteUint64(6)
	WriteMem(pathMem)

	ReadMem(dstMem)
	return int64(ReadUint64())
}

func _sa_resource_len(pathMem SAMem) int64 {
	WriteUint64(7)
	WriteMem(pathMem)
	return int64(ReadUint64())
}

func _sa_sql_write(dbMem SAMem, queryMem SAMem) int64 {
	WriteUint64(8)
	WriteMem(dbMem)
	WriteMem(queryMem)
	return int64(ReadUint64())
}

func _sa_sql_read(dbMem SAMem, queryMem SAMem) int64 {
	WriteUint64(9)
	WriteMem(dbMem)
	WriteMem(queryMem)
	return int64(ReadUint64())
}

func _sa_sql_readRowLen(dbMem SAMem, queryMem SAMem, queryHash int64, row_i uint64) int64 {
	WriteUint64(10)
	WriteMem(dbMem)
	WriteMem(queryMem)
	WriteUint64(uint64(queryHash))
	WriteUint64(row_i)
	return int64(ReadUint64())
}

func _sa_sql_readRow(dbMem SAMem, queryMem SAMem, queryHash int64, row_i uint64, resultMem SAMem) int64 {
	WriteUint64(11)
	WriteMem(dbMem)
	WriteMem(queryMem)
	WriteUint64(uint64(queryHash))
	WriteUint64(row_i)

	ReadMem(resultMem)
	return int64(ReadUint64())
}

func _sa_div_colResize(pos uint64, nameMem SAMem, val float64) float64 {
	WriteUint64(12)
	WriteUint64(pos)
	WriteMem(nameMem)
	WriteFloat64(val)

	return ReadFloat64()
}
func _sa_div_rowResize(pos uint64, nameMem SAMem, val float64) float64 {
	WriteUint64(13)
	WriteUint64(pos)
	WriteMem(nameMem)
	WriteFloat64(val)

	return ReadFloat64()
}
func _sa_div_colMax(pos uint64, val float64) float64 {
	WriteUint64(14)
	WriteUint64(pos)
	WriteFloat64(val)

	return ReadFloat64()
}

func _sa_div_rowMax(pos uint64, val float64) float64 {
	WriteUint64(15)
	WriteUint64(pos)
	WriteFloat64(val)

	return ReadFloat64()
}

func _sa_div_col(pos uint64, val float64) float64 {
	WriteUint64(16)
	WriteUint64(pos)
	WriteFloat64(val)

	return ReadFloat64()
}

func _sa_div_row(pos uint64, val float64) float64 {
	WriteUint64(17)
	WriteUint64(pos)
	WriteFloat64(val)

	return ReadFloat64()
}

func _sa_div_start(x, y, w, h uint64, nameMem SAMem) int64 {
	WriteUint64(18)
	WriteUint64(x)
	WriteUint64(y)
	WriteUint64(w)
	WriteUint64(h)
	WriteMem(nameMem)

	return int64(ReadUint64())
}

func _sa_div_end() {
	WriteUint64(19)
}

//20 can be use later ...

func _sa_div_dialogClose() {
	WriteUint64(21)
}

func _sa_div_dialogStart(nameMem SAMem, tp uint64, openIt uint64) int64 {
	WriteUint64(22)
	WriteMem(nameMem)
	WriteUint64(tp)
	WriteUint64(openIt)

	return int64(ReadUint64())
}

func _sa_div_dialogEnd() {
	WriteUint64(23)
}

func _sa_div_get_info(idMem SAMem, x int64, y int64) float64 {
	WriteUint64(24)
	WriteMem(idMem)
	WriteUint64(uint64(x))
	WriteUint64(uint64(y))

	return ReadFloat64()
}

func _sa_div_set_info(idMem SAMem, val float64, x int64, y int64) float64 {
	WriteUint64(25)
	WriteMem(idMem)
	WriteFloat64(val)
	WriteUint64(uint64(x))
	WriteUint64(uint64(y))

	return ReadFloat64()
}

func _sa_paint_rect(x, y, w, h float64, margin float64, r, g, b, a uint32, borderWidth float64) int64 {
	WriteUint64(26)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteFloat64(margin)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteFloat64(borderWidth)

	return int64(ReadUint64())
}

func _sa_paint_line(x, y, w, h float64, margin float64, sx, sy, ex, ey float64, r, g, b, a uint32, width float64) int64 {
	WriteUint64(27)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteFloat64(margin)
	WriteFloat64(sx)
	WriteFloat64(sy)
	WriteFloat64(ex)
	WriteFloat64(ey)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteFloat64(width)

	return int64(ReadUint64())
}

func _sa_paint_circle(x, y, w, h float64, margin float64, sx, sy, rad float64, r, g, b, a uint32, borderWidth float64) int64 {
	WriteUint64(28)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteFloat64(margin)
	WriteFloat64(sx)
	WriteFloat64(sy)
	WriteFloat64(rad)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteFloat64(borderWidth)

	return int64(ReadUint64())
}

func _sa_paint_file(x, y, w, h float64, fileMem SAMem, titleMem SAMem, margin, marginX, marginY float64, r, g, b, a uint32, alignV, alignH uint32, fill, inverse uint32) int64 {
	WriteUint64(29)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteMem(fileMem)
	WriteMem(titleMem)
	WriteFloat64(margin)
	WriteFloat64(marginX)
	WriteFloat64(marginY)
	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))
	WriteUint64(uint64(alignV))
	WriteUint64(uint64(alignH))
	WriteUint64(uint64(fill))
	WriteUint64(uint64(inverse))

	return int64(ReadUint64())
}

func _sa_paint_text(x, y, w, h float64,
	valueMem SAMem,
	margin float64, marginX float64, marginY float64,
	r, g, b, a uint32,
	ratioH, lineHeight float64,
	fontId, align, alignV uint32,
	selection, edit, tabIsChar, enable uint32) int64 {

	WriteUint64(30)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)

	WriteMem(valueMem)

	WriteFloat64(margin)
	WriteFloat64(marginX)
	WriteFloat64(marginY)

	WriteUint64(uint64(r))
	WriteUint64(uint64(g))
	WriteUint64(uint64(b))
	WriteUint64(uint64(a))

	WriteFloat64(ratioH)
	WriteFloat64(lineHeight)

	WriteUint64(uint64(fontId))
	WriteUint64(uint64(align))
	WriteUint64(uint64(alignV))
	WriteUint64(uint64(selection))
	WriteUint64(uint64(edit))
	WriteUint64(uint64(tabIsChar))
	WriteUint64(uint64(enable))

	return int64(ReadUint64())
}

func _sa_paint_textWidth(valueMem SAMem, fontId uint32, ratioH float64, cursorPos int64) float64 {
	WriteUint64(31)
	WriteMem(valueMem)
	WriteUint64(uint64(fontId))
	WriteFloat64(ratioH)
	WriteUint64(uint64(cursorPos))

	return ReadFloat64()
}

func _sa_paint_title(x, y, w, h float64, valueMem SAMem) int64 {
	WriteUint64(32)
	WriteFloat64(x)
	WriteFloat64(y)
	WriteFloat64(w)
	WriteFloat64(h)
	WriteMem(valueMem)

	return int64(ReadUint64())
}

func _sa_paint_cursor(nameMem SAMem) int64 {
	WriteUint64(33)
	WriteMem(nameMem)

	return int64(ReadUint64())
}

func _sa_print(mem SAMem) {
	WriteUint64(34)
	WriteMem(mem)
}

func _sa_print_float(val float64) {
	WriteUint64(35)
	WriteFloat64(val)
}

func _sa_fn_call(assetMem SAMem, fnMem SAMem, argsMem SAMem) int64 {
	WriteUint64(36)
	WriteMem(assetMem)
	WriteMem(fnMem)
	WriteMem(argsMem)

	return int64(ReadUint64())
}

func _sa_fn_setReturn(argsMem SAMem) int64 {
	WriteUint64(37)
	WriteMem(argsMem)

	return int64(ReadUint64())
}

func _sa_fn_getReturn(argsMem SAMem) int64 {
	WriteUint64(38)

	ReadMem(argsMem)
	return int64(ReadUint64())
}

func _sa_swp_drawButton(cd_r, cd_g, cd_b, cd_a uint32,
	frontCd_r, frontCd_g, frontCd_b, frontCd_a uint32,
	valueMem SAMem, iconMem SAMem, urlMem SAMem, titleMem SAMem,
	font uint32, alpha float64, alphaNoBack uint32, iconInverseColor uint32,
	margin float64, marginIcon float64, align uint32, ratioH float64,
	enable uint32, highlight uint32, drawBorder uint32,
	outMem SAMem) int64 {

	WriteUint64(39)

	WriteUint64(uint64(cd_r))
	WriteUint64(uint64(cd_g))
	WriteUint64(uint64(cd_b))
	WriteUint64(uint64(cd_a))

	WriteUint64(uint64(frontCd_r))
	WriteUint64(uint64(frontCd_g))
	WriteUint64(uint64(frontCd_b))
	WriteUint64(uint64(frontCd_a))

	WriteMem(valueMem)
	WriteMem(iconMem)
	WriteMem(urlMem)
	WriteMem(titleMem)

	WriteUint64(uint64(font))
	WriteFloat64(alpha)
	WriteUint64(uint64(alphaNoBack))
	WriteUint64(uint64(iconInverseColor))

	WriteFloat64(margin)
	WriteFloat64(marginIcon)
	WriteUint64(uint64(align))
	WriteFloat64(ratioH)

	WriteUint64(uint64(enable))
	WriteUint64(uint64(highlight))
	WriteUint64(uint64(drawBorder))

	ReadMem(outMem)
	return int64(ReadUint64())
}

func _sa_swp_drawSlider(value float64, min float64, max float64, jump float64, titleMem SAMem, enable uint32, outMem SAMem) float64 {
	WriteUint64(40)
	WriteFloat64(value)
	WriteFloat64(min)
	WriteFloat64(max)
	WriteFloat64(jump)
	WriteMem(titleMem)
	WriteUint64(uint64(enable))

	ReadMem(outMem)
	return ReadFloat64()
}

func _sa_swp_drawProgress(value float64, maxValue float64, titleMem SAMem, margin float64, enable uint32) int64 {
	WriteUint64(41)
	WriteFloat64(value)
	WriteFloat64(maxValue)
	WriteMem(titleMem)
	WriteFloat64(margin)
	WriteUint64(uint64(enable))
	return int64(ReadUint64())
}

func _sa_swp_drawText(cd_r, cd_g, cd_b, cd_a uint32,
	valueMem SAMem, titleMem SAMem, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, alignV uint32, ratioH float64,
	enable uint32, selection uint32) int64 {
	WriteUint64(42)
	WriteUint64(uint64(cd_r))
	WriteUint64(uint64(cd_g))
	WriteUint64(uint64(cd_b))
	WriteUint64(uint64(cd_a))

	WriteMem(valueMem)
	WriteMem(titleMem)
	WriteUint64(uint64(font))

	WriteFloat64(margin)
	WriteFloat64(marginX)
	WriteFloat64(marginY)
	WriteUint64(uint64(align))
	WriteUint64(uint64(alignV))
	WriteFloat64(ratioH)

	WriteUint64(uint64(enable))
	WriteUint64(uint64(selection))

	return int64(ReadUint64())
}

func _sa_swp_getEditValue(outMem SAMem) int64 {
	WriteUint64(43)

	ReadMem(outMem)
	return int64(ReadUint64())
}

func _sa_swp_drawEdit(cd_r, cd_g, cd_b, cd_a uint32,
	valueMem SAMem, valueOrigMem SAMem, titleMem SAMem, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, alignV uint32, ratioH float64,
	enable uint32,
	outMem SAMem) int64 {
	WriteUint64(44)
	WriteUint64(uint64(cd_r))
	WriteUint64(uint64(cd_g))
	WriteUint64(uint64(cd_b))
	WriteUint64(uint64(cd_a))

	WriteMem(valueMem)
	WriteMem(valueOrigMem)
	WriteMem(titleMem)
	WriteUint64(uint64(font))

	WriteFloat64(margin)
	WriteFloat64(marginX)
	WriteFloat64(marginY)
	WriteUint64(uint64(align))
	WriteUint64(uint64(alignV))
	WriteFloat64(ratioH)

	WriteUint64(uint64(enable))

	ReadMem(outMem)
	return int64(ReadUint64())
}

func _sa_swp_drawCombo(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, optionsMem SAMem, titleMem SAMem, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, ratioH float64,
	enable uint32) int64 {
	WriteUint64(45)
	WriteUint64(uint64(cd_r))
	WriteUint64(uint64(cd_g))
	WriteUint64(uint64(cd_b))
	WriteUint64(uint64(cd_a))

	WriteUint64(value)
	WriteMem(optionsMem)
	WriteMem(titleMem)
	WriteUint64(uint64(font))

	WriteFloat64(margin)
	WriteFloat64(marginX)
	WriteFloat64(marginY)
	WriteUint64(uint64(align))
	WriteFloat64(ratioH)

	WriteUint64(uint64(enable))

	return int64(ReadUint64())
}

func _sa_swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, descriptionMem SAMem, titleMem SAMem,
	height float64, align uint32, alignV uint32, enable uint32) int64 {
	WriteUint64(46)
	WriteUint64(uint64(cd_r))
	WriteUint64(uint64(cd_g))
	WriteUint64(uint64(cd_b))
	WriteUint64(uint64(cd_a))

	WriteUint64(value)
	WriteMem(descriptionMem)
	WriteMem(titleMem)

	WriteFloat64(height)
	WriteUint64(uint64(align))
	WriteUint64(uint64(alignV))
	WriteUint64(uint64(enable))

	return int64(ReadUint64())
}

func _sa_div_drag(groupName SAMem, id uint64) int64 {
	WriteUint64(47)
	WriteMem(groupName)
	WriteUint64(id)
	return int64(ReadUint64())
}

func _sa_div_drop(groupName SAMem, vertical uint32, horizontal uint32, inside uint32, outMem SAMem) int64 {
	WriteUint64(48)
	WriteMem(groupName)
	WriteUint64(uint64(vertical))
	WriteUint64(uint64(horizontal))
	WriteUint64(uint64(inside))

	ReadMem(outMem)
	return int64(ReadUint64())
}

func _sa_render_app(appMem SAMem, dbMem SAMem, sts_id uint64) int64 {
	WriteUint64(49)
	WriteMem(appMem)
	WriteMem(dbMem)
	WriteUint64(sts_id)

	return int64(ReadUint64())
}

/*func _SA_ptrToString(mem SAMem) string {
	ptr := uint32(mem >> 32)
	size := uint32(mem)
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), size)
}*/

type SAMem struct {
	v []byte
}

func _SA_stringToPtr(s string) SAMem {
	return SAMem{v: []byte(s)}
}

func _SA_bytesToPtr(s []byte) SAMem {
	return SAMem{v: s}
}
func _SA_ptrToBytes(mem SAMem) []byte {
	return mem.v
}
