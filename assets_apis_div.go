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
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
)

func (asset *Asset) _VmBasic_touchScrollEnabled(packLayout *LayoutDiv) (bool, bool) {

	root := asset.app.root

	insideScrollV := false
	insideScrollH := false
	if packLayout.data.scrollV.Is() {
		scrollQuad := packLayout.data.scrollV.GetScrollBackCoordV(packLayout.crop, root.ui)
		insideScrollV = scrollQuad.Inside(root.ui.io.touch.pos)
	}

	if packLayout.data.scrollH.Is() {
		scrollQuad := packLayout.data.scrollH.GetScrollBackCoordH(packLayout.crop, root.ui)
		insideScrollH = scrollQuad.Inside(root.ui.io.touch.pos)
	}
	return insideScrollV, insideScrollH
}

func (asset *Asset) _VmBasic_touchScroll(packLayout *LayoutDiv, enableInput bool) {

	root := asset.app.root

	hasScrollV := packLayout.data.scrollV.Is()
	hasScrollH := packLayout.data.scrollH.Is()

	if hasScrollV {
		scrollQuad := packLayout.data.scrollV.GetScrollBackCoordV(packLayout.crop, root.ui)
		if enableInput {
			if packLayout.data.scrollV.TouchV(packLayout.CropWithScroll(root.ui), scrollQuad, root.touch.IsFnMove(nil, packLayout, nil, nil), root) {
				root.touch.Set(nil, packLayout, nil, nil)
			}
		}
	} //else {
	//packLayout.data.scrollV.SetWheel(packLayout.data.scrollV.GetWheel())
	//}

	if hasScrollH {
		scrollQuad := packLayout.data.scrollH.GetScrollBackCoordH(packLayout.crop, root.ui)
		if enableInput {
			if packLayout.data.scrollH.TouchH(packLayout.CropWithScroll(root.ui), scrollQuad, hasScrollV, root.touch.IsFnMove(nil, nil, packLayout, nil), root) {
				root.touch.Set(nil, nil, packLayout, nil)
			}
		}
	} //else {
	//packLayout.data.scrollH.SetWheel(packLayout.data.scrollH.GetWheel())
	//}
}

func (asset *Asset) _VmBasic_RenderScroll(packLayout *LayoutDiv, showBackground bool, buff *PaintBuff) {

	if packLayout.data.scrollV.Is() {
		scrollQuad := packLayout.data.scrollV.GetScrollBackCoordV(packLayout.crop, buff.ui)
		packLayout.data.scrollV.DrawV(scrollQuad, showBackground, buff)
	}

	if packLayout.data.scrollH.Is() {
		scrollQuad := packLayout.data.scrollH.GetScrollBackCoordH(packLayout.crop, buff.ui)
		packLayout.data.scrollH.DrawH(scrollQuad, showBackground, buff)
	}
}

func (asset *Asset) renderStart() {

	root := asset.app.root
	st := root.stack

	st.stack.data.Reset() //here because after *dialog* needs to know old size
	st.stack.UpdateCoord(root.ui)

	enableInput := st.stack.data.touch_enabled
	if st.stack.parent != nil {
		enableInput = enableInput && st.stack.parent.enableInput
	}
	asset._VmBasic_touchScroll(st.stack, enableInput)

	// scroll touch
	insideScrollV, insideScrollH := asset._VmBasic_touchScrollEnabled(st.stack)
	overScroll := enableInput && (insideScrollV || insideScrollH)
	enableInput = enableInput && !insideScrollV && !insideScrollH //can NOT click through

	startTouch := enableInput && root.ui.io.touch.start && !root.ui.io.keys.alt
	endTouch := enableInput && root.ui.io.touch.end

	over := enableInput && st.stack.crop.Inside(root.ui.io.touch.pos)
	inside := over
	if inside && startTouch && enableInput {
		if !root.touch.IsScrollOrResizeActive() { //if lower resize or scroll is activated than don't rewrite it with higher canvas
			root.touch.Set(st.stack, nil, nil, nil)
		}
	}
	touchActiveMove := root.touch.IsFnMove(st.stack, nil, nil, nil)

	if !touchActiveMove && enableInput && root.touch.IsAnyActive() { // when click and move, other Buttons, etc. are disabled
		inside = false
	}

	st.stack.enableInput = enableInput

	st.stack.data.over = over
	st.stack.data.overScroll = overScroll
	st.stack.data.touch_inside = inside
	st.stack.data.touch_active = touchActiveMove
	st.stack.data.touch_end = (endTouch && touchActiveMove) //&& inside

	/*if st.stack.parent != nil {
		st.stack.parent.gridLock = true
	}*/

	st.buff.AddCrop(st.stack.crop)

	if root.ui.io.ini.Grid {
		asset.DrawGrid()
	}
}

func (asset *Asset) DrawGrid() {
	root := asset.app.root
	st := root.stack

	start := st.stack.canvas.Start
	size := st.stack.canvas.Size

	cd := OsCd{200, 100, 80, 200}

	//cols
	px := start.X
	for _, col := range st.stack.data.cols.outputs {
		px += int(col)
		st.buff.AddLine(OsV2{px, start.Y}, OsV2{px, start.Y + size.Y}, cd, asset.getCellWidth(0.03))
	}

	//rows
	py := start.Y
	for _, row := range st.stack.data.rows.outputs {
		py += int(row)
		st.buff.AddLine(OsV2{start.X, py}, OsV2{start.X + size.X, py}, cd, asset.getCellWidth(0.03))
	}

	px = start.X
	for x, col := range st.stack.data.cols.outputs {

		py = start.Y
		for y, row := range st.stack.data.rows.outputs {
			st.buff.AddText(fmt.Sprintf("[%d, %d]", x, y), OsV4{OsV2{px, py}, OsV2{int(col), int(row)}}, root.fonts.Get(0), cd, root.ui.io.GetDPI()/8, OsV2{1, 1}, nil)
			py += int(row)
		}

		px += int(col)
	}

}

func (asset *Asset) renderEnd(baseDiv bool) {

	root := asset.app.root
	st := root.stack

	st.stack.gridLock = false

	// show scroll
	st.buff.AddCrop(st.stack.CropWithScroll(root.ui))
	asset._VmBasic_RenderScroll(st.stack, st.stack.data.scrollOnScreen, st.buff)

	if st.stack.parent != nil {
		st.stack = st.stack.parent
		st.buff.AddCrop(st.stack.crop)
	} else {
		if !baseDiv {
			asset.AddLogErr(fmt.Errorf("div==nil in level: %s. Check if every 'start' has 'end'. Check return/continue/break in the middle of 'start' - 'end'", st.name))
		}
	}
}

func (asset *Asset) div_start(x, y, w, h uint64, name string) int64 {

	root := asset.app.root
	st := root.stack

	if !st.stack.gridLock {

		// cols/rows resizer
		st.stack.RenderResizeSpliter(root, st.buff)

		st.stack.UpdateGrid(root.ui)
		st.stack.lastChild = nil
		st.stack.gridLock = true
	}

	grid := InitOsQuad(int(x), int(y), int(w), int(h))
	st.stack = st.stack.FindOrCreate(name, grid, &root.infoLayout)

	asset.renderStart()

	return int64(OsTrn(!st.stack.crop.IsZero(), 1, 0))
}

func (asset *Asset) _sa_div_start(x, y, w, h uint64, nameMem uint64) int64 {
	name, err := asset.ptrToString(nameMem)
	if asset.AddLogErr(err) {
		return -1
	}
	return asset.div_start(x, y, w, h, name)
}

func (asset *Asset) _sa_div_end() {
	asset.renderEnd(false)
}

func (asset *Asset) checkGridLock() bool {

	if asset.app.root.stack.stack.gridLock {
		fmt.Println("Warning: Trying to changed col/row dimension after you already draw div into")
		return false
	}
	return true
}

func (asset *Asset) _sa_div_col(pos uint64, val float64) float64 {
	if !asset.checkGridLock() {
		return -1
	}

	root := asset.app.root
	st := root.stack
	st.stack.GetInputCol(int(pos)).min = float32(val)

	return float64(st.stack.data.cols.GetOutput(int(pos))) / float64(root.ui.Cell())
}

func (asset *Asset) div_colResize(pos uint64, name string, val float64) float64 {
	if !asset.checkGridLock() {
		return -1
	}

	root := asset.app.root
	st := root.stack

	//if 'resize' exist in layout than read it from there
	if len(name) == 0 {
		name = strconv.Itoa(int(pos))
	}

	res, found := st.stack.data.cols.FindOrAddResize(name)
	if !found {
		res.value = float32(val)
	}
	st.stack.GetInputCol(int(pos)).resize = res

	//st.stack.GetInputCol(int(pos)).resize = float32(val)
	//st.stack.GetInputCol(int(pos)).resize_name = name

	return float64(st.stack.data.cols.GetOutput(int(pos))) / float64(root.ui.Cell())
}

func (asset *Asset) div_rowResize(pos uint64, name string, val float64) float64 {
	if !asset.checkGridLock() {
		return -1
	}

	root := asset.app.root
	st := root.stack

	//if 'resize' exist in layout than read it from there
	res, found := st.stack.data.rows.FindOrAddResize(name)
	if !found {
		res.value = float32(val)
	}
	st.stack.GetInputRow(int(pos)).resize = res

	//st.stack.GetInputRow(int(pos)).resize_name = name

	return float64(st.stack.data.rows.GetOutput(int(pos))) / float64(root.ui.Cell())
}

func (asset *Asset) _sa_div_colResize(pos uint64, nameMem uint64, val float64) float64 {
	name, err := asset.ptrToString(nameMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.div_colResize(pos, name, val)
}

func (asset *Asset) _sa_div_colMax(pos uint64, val float64) float64 {
	if !asset.checkGridLock() {
		return -1
	}

	root := asset.app.root
	st := root.stack

	st.stack.GetInputCol(int(pos)).max = float32(val)

	return float64(st.stack.data.cols.GetOutput(int(pos))) / float64(root.ui.Cell())
}

func (asset *Asset) _sa_div_row(pos uint64, val float64) float64 {
	if !asset.checkGridLock() {
		return -1
	}

	root := asset.app.root
	st := root.stack
	st.stack.GetInputRow(int(pos)).min = float32(val)

	return float64(st.stack.data.rows.GetOutput(int(pos))) / float64(root.ui.Cell())
}

func (asset *Asset) _sa_div_rowResize(pos uint64, nameMem uint64, val float64) float64 {

	name, err := asset.ptrToString(nameMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.div_rowResize(pos, name, val)
}

func (asset *Asset) _sa_div_rowMax(pos uint64, val float64) float64 {
	if !asset.checkGridLock() {
		return -1
	}

	root := asset.app.root
	st := root.stack

	st.stack.GetInputRow(int(pos)).max = float32(val)

	return float64(st.stack.data.rows.GetOutput(int(pos))) / float64(root.ui.Cell())
}

func (asset *Asset) _sa_div_dialogClose() {
	asset.app.root.stack.close = true
}

func (asset *Asset) div_dialogStart(name string, tp uint64, openIt bool) int64 {
	root := asset.app.root
	st := asset.app.root.stack

	//name
	if len(name) == 0 {
		return -1
	}

	//coord
	var src_coordMoveCut OsV4
	switch tp {
	case 1:
		if st.stack.lastChild != nil {
			src_coordMoveCut = st.stack.lastChild.crop
		} else {
			src_coordMoveCut = st.stack.crop
		}
	case 2:
		src_coordMoveCut = OsV4{Start: root.ui.io.touch.pos, Size: OsV2{1, 1}}
	}

	if openIt {
		//add
		st.AddLevel(name, src_coordMoveCut, st.stack, root.ui)

		root.touch.Reset()
		root.ui.io.ResetTouchAndKeys()
		root.ui.io.edit.setFirstEditbox = true
	}

	lev := st.next
	if lev != nil && lev.name == name && lev.openPack == st.stack {
		lev.use = true
		if tp < 2 { //not for touch_pos
			lev.src_coordMoveCut = src_coordMoveCut
		}

		//coord
		winRect, _ := root.ui.GetScreenCoord()

		coord := lev.div.GetLevelSize(winRect, root.ui)
		coord = lev.GetCoord(coord, winRect)
		lev.div.canvas = coord
		lev.div.crop = coord

		root.SetLevel(lev)
		lev.buff.Reset(lev.stack.canvas)

		//fade
		lev.buff.AddCrop(winRect)
		lev.buff.AddRect(winRect, OsCd{0, 0, 0, 80}, 0)
		//background
		lev.buff.AddCrop(lev.stack.canvas)
		lev.buff.AddRect(lev.stack.canvas, OsCd_white(), 0)

		asset.renderStart()

		return 1 //active
	}

	return 0 //not opened
}

func (asset *Asset) _sa_div_dialogStart(nameMem uint64, tp uint64, openIt uint64) int64 {

	name, err := asset.ptrToString(nameMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.div_dialogStart(name, tp, openIt > 0)
}

func (asset *Asset) _sa_div_dialogEnd() {

	root := asset.app.root
	st := root.stack

	parent := st.parent
	if parent == nil {
		return
	}

	//close dialog
	if st.stack.enableInput {
		winRect, _ := root.ui.GetScreenCoord()
		outside := winRect.Inside(root.ui.io.touch.pos) && !st.div.canvas.Inside(root.ui.io.touch.pos)
		if (root.ui.io.touch.end && outside) || root.ui.io.keys.esc {
			root.touch.Reset()
			st.close = true
			root.ui.io.keys.esc = false
		}
	}

	asset.renderEnd(true)

	root.stack = parent
}

func (asset *Asset) div_get_info(id string, x int64, y int64) float64 {

	root := asset.app.root
	st := root.stack

	div := st.stack
	if div != nil && (x >= 0 || y >= 0) {
		div = div.FindInside(OsV2{int(x), int(y)})
	}
	if div == nil {
		return -1
	}

	switch id {
	case "cell":
		return float64(root.ui.Cell())

	case "layoutWidth":
		return float64(div.canvas.Size.X) / float64(root.ui.Cell())
	case "layoutHeight":
		return float64(div.canvas.Size.Y) / float64(root.ui.Cell())

	case "screenWidth":
		return float64(div.crop.Size.X) / float64(root.ui.Cell())
	case "screenHeight":
		return float64(div.crop.Size.Y) / float64(root.ui.Cell())

	case "layoutStartX":
		return float64(div.crop.Start.X-div.canvas.Start.X) / float64(root.ui.Cell())
	case "layoutStartY":
		return float64(div.crop.Start.Y-div.canvas.Start.Y) / float64(root.ui.Cell())

	case "touch":
		return OsTrnFloat(div.enableInput, 1, 0)

	case "touchX":
		rpos := OsV2{-1, -1}
		if div.enableInput {
			rpos = root.ui.io.touch.pos.Sub(div.canvas.Start)
		}
		return float64(rpos.X) / float64(div.canvas.Size.X)
	case "touchY":
		rpos := OsV2{-1, -1}
		if div.enableInput {
			rpos = root.ui.io.touch.pos.Sub(div.canvas.Start)
		}
		return float64(rpos.Y) / float64(div.canvas.Size.Y)

	case "touchOver":
		return OsTrnFloat(div.data.over, 1, 0)

	case "touchOverScroll":
		return OsTrnFloat(div.data.overScroll, 1, 0)

	case "touchInside":
		return OsTrnFloat(div.data.touch_inside, 1, 0)

	case "touchStart":
		if div.enableInput {
			return OsTrnFloat(root.ui.io.touch.start, 1, 0)
		} else {
			return 0
		}
	case "touchWheel":
		if div.enableInput {
			return float64(root.ui.io.touch.wheelPos)
		} else {
			return 0
		}
	case "touchClicks":
		if div.enableInput {
			return float64(root.ui.io.touch.numClicks)
		} else {
			return 0
		}
	case "touchForce":
		if div.enableInput {
			return OsTrnFloat(root.ui.io.touch.rm, 1, 0)
		} else {
			return 0
		}

	case "touchActive":
		return OsTrnFloat(div.data.touch_active, 1, 0)
	case "touchEnd":
		return OsTrnFloat(div.data.touch_end, 1, 0)

	case "scrollVpos":
		return float64(div.data.scrollV.GetWheel()) / float64(root.ui.Cell())
	case "scrollHpos":
		return float64(div.data.scrollH.GetWheel()) / float64(root.ui.Cell())

	case "scrollVshow":
		return OsTrnFloat(div.data.scrollV.show, 1, 0)
	case "scrollHshow":
		return OsTrnFloat(div.data.scrollH.show, 1, 0)

	case "scrollVnarrow":
		return OsTrnFloat(div.data.scrollV.narrow, 1, 0)
	case "scrollHnarrow":
		return OsTrnFloat(div.data.scrollH.narrow, 1, 0)
	}

	return -1
}

func (asset *Asset) div_set_info(id string, val float64, x int64, y int64) float64 {

	root := asset.app.root
	st := root.stack

	div := st.stack
	if div != nil && (x >= 0 || y >= 0) {
		div = div.FindInside(OsV2{int(x), int(y)})
	}
	if div == nil {
		return -1
	}

	switch id {

	case "touch_enable":
		bck := div.data.touch_enabled
		div.data.touch_enabled = OsTrnBool(val > 0, true, false)
		return OsTrnFloat(bck, 1, 0)

	case "scrollVpos":
		bck := float64(div.data.scrollV.GetWheel()) / float64(root.ui.Cell())
		div.data.scrollV.wheel = int(val * float64(root.ui.Cell()))
		return bck

	case "scrollHpos":
		bck := float64(div.data.scrollH.GetWheel()) / float64(root.ui.Cell())
		div.data.scrollH.wheel = int(val * float64(root.ui.Cell()))
		return bck

	case "scrollVshow":
		bck := div.data.scrollV.show
		div.data.scrollV.show = OsTrnBool(val > 0, true, false)
		return OsTrnFloat(bck, 1, 0)

	case "scrollHshow":
		bck := div.data.scrollH.show
		div.data.scrollH.show = OsTrnBool(val > 0, true, false)
		return OsTrnFloat(bck, 1, 0)

	case "scrollVnarrow":
		bck := div.data.scrollV.narrow
		div.data.scrollV.narrow = OsTrnBool(val > 0, true, false)
		return OsTrnFloat(bck, 1, 0)

	case "scrollHnarrow":
		bck := div.data.scrollH.narrow
		div.data.scrollH.narrow = OsTrnBool(val > 0, true, false)
		return OsTrnFloat(bck, 1, 0)

	case "scrollOnScreen":
		bck := div.data.scrollOnScreen
		div.data.scrollOnScreen = OsTrnBool(val > 0, true, false)
		return OsTrnFloat(bck, 1, 0)

	}

	return -1
}

func (asset *Asset) _sa_div_get_info(idMem uint64, x int64, y int64) float64 {

	id, err := asset.ptrToString(idMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.div_get_info(id, x, y)
}

func (asset *Asset) _sa_div_set_info(idMem uint64, val float64, x int64, y int64) float64 {

	id, err := asset.ptrToString(idMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.div_set_info(id, val, x, y)
}

func (asset *Asset) div_drag(groupName string, id uint64) int64 {

	root := asset.app.root
	st := root.stack

	if st.stack.data.touch_active {
		drag := &root.ui.io.drag
		//set
		drag.div = root.stack.stack
		drag.group = groupName
		drag.id = id

		//paint
		asset.paint_rect(0, 0, 1, 1, 0, OsCd{0, 0, 0, 180}, 0) //fade
	}
	return 1
}
func (asset *Asset) div_drop(groupName string, vertical uint32, horizontal uint32, inside uint32) (uint64, uint64, int64) {

	root := asset.app.root
	st := root.stack

	id := uint64(0)
	pos := uint64(0)
	done := int64(0)

	touchPos := root.ui.io.touch.pos
	q := st.stack.crop

	drag := &root.ui.io.drag
	if q.Inside(touchPos) && strings.EqualFold(drag.group, groupName) && st.stack != drag.div {

		r := touchPos.Sub(st.stack.crop.Middle())

		if vertical > 0 && horizontal > 0 {
			arx := float32(OsAbs(r.X)) / float32(st.stack.crop.Size.X)
			ary := float32(OsAbs(r.Y)) / float32(st.stack.crop.Size.Y)
			if arx > ary {
				if r.X < 0 {
					pos = 3 //H_LEFT
				} else {
					pos = 4 //H_RIGHT
				}
			} else {
				if r.Y < 0 {
					pos = 1 //V_LEFT
				} else {
					pos = 2 //V_RIGHT
				}
			}
		} else if vertical > 0 {
			if r.Y < 0 {
				pos = 1 //V_LEFT
			} else {
				pos = 2 //V_RIGHT
			}
		} else if horizontal > 0 {
			if r.X < 0 {
				pos = 3 //H_LEFT
			} else {
				pos = 4 //H_RIGHT
			}
		}

		if inside > 0 {
			if vertical > 0 {
				q = q.AddSpaceY(st.stack.crop.Size.Y / 3)
			}
			if horizontal > 0 {
				q = q.AddSpaceX(st.stack.crop.Size.X / 3)
			}

			if vertical == 0 && horizontal == 0 {
				pos = 0
			} else if q.Inside(touchPos) {
				pos = 0
			}
		}

		//paint
		wx := float64(asset.getCellWidth(0.1)) / float64(st.stack.canvas.Size.X)
		wy := float64(asset.getCellWidth(0.1)) / float64(st.stack.canvas.Size.Y)
		switch pos {
		case 0: //SA_Drop_INSIDE
			asset.paint_rect(0, 0, 1, 1, 0, OsCd{0, 0, 0, 180}, 0.03)

		case 1: //SA_Drop_V_LEFT
			asset.paint_rect(0, 0, 1, wy, 0, OsCd{0, 0, 0, 180}, 0)

		case 2: //SA_Drop_V_RIGHT
			asset.paint_rect(0, 1-wy, 1, 1, 0, OsCd{0, 0, 0, 180}, 0)

		case 3: //SA_Drop_H_LEFT
			asset.paint_rect(0, 0, wx, 1, 0, OsCd{0, 0, 0, 180}, 0)

		case 4: //SA_Drop_H_RIGHT
			asset.paint_rect(1-wx, 0, 1, 1, 0, OsCd{0, 0, 0, 180}, 0)
		}

		id = drag.id
		//if st.stack.data.touch_end {
		if root.ui.io.touch.end {
			done = 1
		}

	}

	return id, pos, done
}

func (asset *Asset) _sa_div_drag(groupNameMem uint64, id uint64) int64 {

	groupName, err := asset.ptrToString(groupNameMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.div_drag(groupName, id)
}
func (asset *Asset) _sa_div_drop(groupNameMem uint64, vertical uint32, horizontal uint32, inside uint32, outMem uint64) int64 {
	groupName, err := asset.ptrToString(groupNameMem)
	if asset.AddLogErr(err) {
		return -1
	}

	id, pos, done := asset.div_drop(groupName, vertical, horizontal, inside)

	out, err := asset.ptrToBytesDirect(outMem)
	if asset.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(id))
	binary.LittleEndian.PutUint64(out[8:], uint64(pos))

	return done
}

func (asset *Asset) render_app(appName string, dbName string, sts_id uint64) (int64, error) {

	root := asset.app.root
	//st := root.stack

	app, err := root.AddApp(appName, dbName, int(sts_id))
	if err != nil {
		return -1, err
	}

	app.Render(false)
	return 1, nil
}

func (asset *Asset) _sa_render_app(appMem uint64, dbMem uint64, sts_id uint64) int64 {

	app, err := asset.ptrToString(appMem)
	if asset.AddLogErr(err) {
		return -1
	}
	db, err := asset.ptrToString(dbMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.render_app(app, db, sts_id)
	asset.AddLogErr(err)
	return ret
}
