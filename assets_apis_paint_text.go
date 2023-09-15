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
	"fmt"
	"strings"
	"unicode/utf8"
)

func (asset *Asset) paint_text(x, y, w, h float64,
	value string, valueOrigEdit string,
	margin float64, marginX float64, marginY float64,
	cd OsCd,
	ratioH, lineHeight float64,
	fontId, alignH, alignV uint32,
	selection, edit, tabIsChar, enable uint32) int64 {

	root := asset.app.root
	st := root.levels.GetStack()
	if st.stack == nil || st.stack.crop.IsZero() {
		return -1
	}

	if enable == 0 {
		cd = OsCd_Aprox(OsCd_white(), cd, 0.3)
	}

	coord := asset.getCoord(x, y, w, h, margin, marginX, marginY)

	// crop
	imgRectBackup := st.buff.AddCrop(asset.addCoordMargin(st.stack.crop, margin, marginX, marginY).GetIntersect(coord))

	//one liner
	active := asset._VmDraw_Text_line(coord, 0, OsV2{utf8.RuneCountInString(value), 0},
		value, valueOrigEdit,
		cd,
		ratioH, lineHeight, margin, marginX,
		fontId, alignH, alignV,
		selection != 0, edit != 0, tabIsChar != 0)

	if active {
		asset._VmDraw_resetKeys(edit != 0)
	}

	// crop back
	st.buff.AddCrop(imgRectBackup)
	return 1
}

func (asset *Asset) _sa_paint_text(x, y, w, h float64,
	valueMem uint64,
	margin float64, marginX float64, marginY float64,
	r, g, b, a uint32,
	ratioH, lineHeight float64,
	fontId, align, alignV uint32,
	selection, edit, tabIsChar, enable uint32) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.paint_text(x, y, w, h,
		value, value,
		margin, marginX, marginY,
		InitOsCd32(r, g, b, a),
		ratioH, lineHeight,
		fontId, align, alignV,
		selection, edit, tabIsChar, enable)
}

func _VmDraw_WordPos(str string, mid int) (int, int) {
	start := 0
	end := 0

	str = strings.ToLower(str)

	p := 0
	for _, ch := range str {

		if OsIsTextWord(ch) {
			end = p + 1
		} else {
			if p < mid {
				start = p + 1
			} else {
				break
			}
		}
		p++
	}
	if end < start {
		end = start
	}

	return start, end
}

func (asset *Asset) _VmDraw_resetKeys(editable bool) {

	keys := &asset.app.root.ui.io.keys

	//copy/cut/paste
	keys.copy = false
	keys.cut = false
	keys.paste = false

	//arrows
	keys.arrowL = false
	keys.arrowR = false
	keys.home = false
	keys.end = false

	if editable {
		keys.text = ""
		keys.delete = false
		keys.backspace = false

		keys.esc = false
	}
}

func (asset *Asset) _VmDraw_Text_VScrollInto(cursor OsV2, lineH int) {

	st := asset.app.root.levels.GetStack()
	if st.stack.parent == nil {
		return
	}

	v_pos := cursor.Y * lineH

	v_st := st.stack.parent.data.scrollV.GetWheel()
	v_sz := st.stack.crop.Size.Y - lineH
	v_en := v_st + v_sz

	if v_pos <= v_st {
		st.stack.parent.data.scrollV.SetWheel(OsMax(0, v_pos))
	} else if v_pos >= v_en {
		st.stack.parent.data.scrollV.wheel = OsMax(0, v_pos-v_sz) //SetWheel() has boundary check, which is not good here
	}
}
func (asset *Asset) _VmDraw_Text_HScrollInto(str string, cursor OsV2, font *Font, textH int, margin float64, marginX float64) error {

	st := asset.app.root.levels.GetStack()
	if st.stack.parent == nil {
		return nil
	}

	h_pos, err := font.GetPxPos(str, textH, cursor.X)
	if err == nil {
		h_align := asset.getCellWidth(margin + marginX) //margin + marginX

		h_st := st.stack.parent.data.scrollH.GetWheel()
		h_sz := st.stack.crop.Size.X - 3*h_align
		h_en := h_st + h_sz

		if h_pos <= h_st {
			st.stack.parent.data.scrollH.SetWheel(OsMax(0, h_pos))
		} else if h_pos >= h_en {
			st.stack.parent.data.scrollH.wheel = OsMax(0, h_pos-h_sz) //SetWheel() has boundary check, which is not good here
		}
	}
	return err
}

func (asset *Asset) _VmDraw_TextSelectTouch(str string, strEditOrig string, touchPos OsV2, lineEnd OsV2, editable bool, font *Font, textH int, lineH int, margin float64, marginX float64) {

	root := asset.app.root
	st := root.levels.GetStack()

	//dict := stt.dict
	edit := &root.ui.io.edit
	keys := &root.ui.io.keys
	touch := &root.ui.io.touch

	this_uid := st.stack //.Hash()
	edit_uid := edit.uid
	next_uid := edit.next_uid

	active := (edit_uid != nil && edit_uid == this_uid)
	activate_next_uid := (this_uid == next_uid)

	if st.stack.enableInput && ((editable && edit.setFirstEditbox) || (editable && edit.tab) || activate_next_uid) {
		//setFirstEditbox or Tab
		edit.uid = this_uid

		if !active {
			edit.temp = strEditOrig
			edit.orig = strEditOrig
		}

		if !activate_next_uid {
			//select all
			edit.start = OsV2{}
			edit.end = lineEnd
		}

		edit.setFirstEditbox = false
		edit.next_uid = nil
		edit.tab = false

		root.ui.SetTextCursorMove()
	} else if st.stack.data.touch_inside && touch.start {
		//click inside
		if !active {
			edit.next_uid = this_uid //set next_uid
		}

		//set end
		edit.end = touchPos

		if !active || !keys.shift {
			//set start
			edit.start = touchPos
		}

		switch touch.numClicks {
		case 2:
			first, last := _VmDraw_WordPos(str, touchPos.X)
			edit.start = OsV2{first, touchPos.Y} //set start
			edit.end = OsV2{last, touchPos.Y}    //set end
		case 3:
			edit.start = OsV2{0, touchPos.Y}                         //set start
			edit.end = OsV2{utf8.RuneCountInString(str), touchPos.Y} //set end
		}
	}

	//keep selecting
	if active && st.stack.data.touch_active && (touch.numClicks != 2 && touch.numClicks != 3) {
		edit.end = touchPos //set end

		//scroll
		asset._VmDraw_Text_VScrollInto(touchPos, lineH)
		asset._VmDraw_Text_HScrollInto(str, touchPos, font, textH, margin, marginX)

		root.ui.SetNoSleep()
	}
}

func subString(s string, rune_start int, rune_end int) (int, int) {

	st := len(s)
	en := len(s)

	p := 0
	//convert rune_pos -> byte_pos
	for i := range s {
		if p == rune_start {
			st = i
		}
		if p == rune_end {
			en = i
			break
		}
		p++
	}
	return st, en
}

func _VmDraw_getStringSubBytePosEx(str string, sx int, ex int) (int, int) {
	//swap
	if sx > ex {
		t := sx
		sx = ex
		ex = t
	}
	return subString(str, int(sx), int(ex))
}
func (asset *Asset) _VmDraw_getStringSubBytePos(str string) (int, int, int, int) {

	root := asset.app.root
	edit := &root.ui.io.edit

	sx := edit.start.X
	ex := edit.end.X

	selFirst := sx
	selLast := ex
	if ex < sx {
		selFirst = ex
		selLast = sx
	}

	x, y := _VmDraw_getStringSubBytePosEx(str, int(sx), int(ex))
	return x, y, selFirst, selLast
}

func (asset *Asset) _VmDraw_TextSelectKeys(str string, lineY int, lineEnd OsV2, editable bool) {

	root := asset.app.root
	keys := &root.ui.io.keys
	//dict := stt.dict
	edit := &root.ui.io.edit

	s := &edit.start
	e := &edit.end

	if editable {
		str = edit.temp
	}
	st, en, _, _ := asset._VmDraw_getStringSubBytePos(str)

	//select all
	if keys.selectAll {
		*s = OsV2{}
		*e = lineEnd
	}

	//copy, cut
	if keys.copy || keys.cut {
		keys.clipboard = str[st:en]
	}

	//shift
	if keys.shift {

		//ctrl
		ex := e.X
		if keys.ctrl {
			if keys.arrowL {
				p := OsMax(ex-1, 0)
				first, _ := _VmDraw_WordPos(str, p)
				if first == p && p > 0 {
					first, _ = _VmDraw_WordPos(str, p-1)
				}
				e.X = first
			}
			if keys.arrowR {
				p := OsMin(ex+1, utf8.RuneCountInString(str))
				_, last := _VmDraw_WordPos(str, p)
				if last == p && p+1 < utf8.RuneCountInString(str) {
					_, last = _VmDraw_WordPos(str, p+1)
				}
				e.X = last
			}
		} else {
			if keys.arrowL {
				e.X = OsMax(ex-1, 0)
			}
			if keys.arrowR {
				e.X = OsMin(ex+1, utf8.RuneCountInString(str))
			}
		}

		//home & end
		if keys.home {
			e.X = 0
		}
		if keys.end {
			e.X = utf8.RuneCountInString(str)
		}
	}
}

func (asset *Asset) _VmDraw_TextEditKeys(tabIsChar bool, font *Font, textH int, lineH int, margin float64, marginX float64) string {

	root := asset.app.root
	//stt := &root.stack
	edit := &root.ui.io.edit
	keys := &root.ui.io.keys

	shiftKey := keys.shift

	uid := edit.uid

	s := &edit.start
	e := &edit.end

	old := *e

	//tempRec := &edit.temp
	str := edit.temp
	st, en, selFirst, selLast := asset._VmDraw_getStringSubBytePos(str)

	//cut/paste(copy() is in selectKeys)
	if keys.cut {
		//remove
		str = str[:st] + str[en:]
		edit.temp = str

		//select
		s.X = selFirst
		e.X = selFirst
	} else if keys.paste {
		//remove old selection
		if st != en {
			str = str[:st] + str[en:]
		}

		//insert
		cb := keys.clipboard
		str = str[:st] + cb + str[st:]
		edit.temp = str

		p := selFirst + utf8.RuneCountInString(cb)
		s.X = p
		e.X = p
	}

	//insert text
	txt := keys.text
	if tabIsChar && keys.tab {
		txt += "\t"
	}
	if len(txt) > 0 {
		//remove old selection
		if st != en {
			str = str[:st] + str[en:]
		}

		//insert
		str = str[:st] + txt + str[st:]
		edit.temp = str

		//cursor
		p := selFirst + utf8.RuneCountInString(txt)
		s.X = p
		e.X = p

		//reset
		keys.text = ""
	}

	//delete/backspace
	if st != en {
		if keys.delete || keys.backspace {
			//remove
			str = str[:st] + str[en:]
			edit.temp = str

			//cursor
			s.X = selFirst
			e.X = selFirst
		}
	} else {
		if keys.delete {
			//remove
			if st < len(str) {
				//removes one letter
				st2, en2 := _VmDraw_getStringSubBytePosEx(str, s.X, s.X+1)
				str = str[:st2] + str[en2:]
				edit.temp = str
			}
		} else if keys.backspace {
			//remove
			if st > 0 {
				//removes one letter
				st2, en2 := _VmDraw_getStringSubBytePosEx(str, s.X-1, s.X)
				str = str[:st2] + str[en2:]
				edit.temp = str

				//select
				p := OsMax(0, selFirst-1)
				s.X = p
				e.X = p
			}
		}
	}

	if !shiftKey {
		//arrows
		if st != en {
			if keys.arrowL {
				//from select -> single start
				s.X = selFirst
				e.X = selFirst
			} else if keys.arrowR {
				//from select -> single end
				s.X = selLast
				e.X = selLast
			}
		} else {
			if keys.ctrl {
				if keys.arrowL {
					p := OsMax(e.X-1, 0)
					first, _ := _VmDraw_WordPos(str, p)
					if first == p && p > 0 {
						first, _ = _VmDraw_WordPos(str, p-1)
					}
					s.X = first
					e.X = first
				}
				if keys.arrowR {
					p := OsMin(e.X+1, utf8.RuneCountInString(str))
					_, last := _VmDraw_WordPos(str, p)
					if last == p && p+1 < utf8.RuneCountInString(str) {
						_, last = _VmDraw_WordPos(str, p+1)
					}
					s.X = last
					e.X = last
				}
			} else {
				if keys.arrowL {
					p := OsMax(0, e.X-1)
					s.X = p
					e.X = p
				} else if keys.arrowR {
					p := OsMin(e.X+1, utf8.RuneCountInString(str))
					s.X = p
					e.X = p
				}
			}
		}

		//home/end
		if keys.home {
			s.X = 0
			e.X = 0
		} else if keys.end {
			p := utf8.RuneCountInString(str)
			s.X = p
			e.X = p
		}
	}

	//history
	{
		//app := stt.GetApp()
		his := VmTextHistoryItem{str: str, cur: e.X}

		root.editbox_history.FindOrAdd(uid, his).AddWithTimeOut(his)

		if keys.backward {
			his = root.editbox_history.FindOrAdd(uid, his).Backward(his)
			edit.temp = his.str
			s.X = his.cur
			e.X = his.cur
		}
		if keys.forward {
			his = root.editbox_history.FindOrAdd(uid, his).Forward()
			edit.temp = his.str
			s.X = his.cur
			e.Y = his.cur
		}
	}

	//scroll
	newPos := *e
	if old.Y != newPos.Y {
		asset._VmDraw_Text_VScrollInto(newPos, lineH)
	}
	if old.X != newPos.X {
		asset._VmDraw_Text_HScrollInto(str, newPos, font, textH, margin, marginX)
	}

	return edit.temp
}

func (asset *Asset) _VmDraw_Text_line(coord OsV4, lineY int, lineEnd OsV2,
	value string, valueOrigEdit string,
	cd OsCd,
	ratioH, lineHeight, margin, marginX float64,
	fontId, alignH, alignV uint32,
	selection, editable, tabIsChar bool) bool {

	root := asset.app.root
	st := root.levels.GetStack()

	align := OsV2{int(alignH), int(alignV)}
	lineH := coord.Size.Y

	if ratioH <= 0 {
		ratioH = 0.35
	}
	textH := asset.getCellWidth(ratioH)

	font := root.fonts.Get(SKYALT_FONT_0) //...int(fontId))
	edit := &root.ui.io.edit
	keys := &root.ui.io.keys
	touch := &root.ui.io.touch

	// mouse pos on text
	touchPos, err := font.GetTextPos(root.ui.io.touch.pos, value, coord, textH, align)
	if err != nil {
		fmt.Println("Error: VmDraw_Text.GetTextPos() failed: %w", err)
		return false
	}

	active := false
	oldCursorPos := edit.end
	cursorPos := OsV2{-1, -1}
	if selection || editable {

		if coord.Inside(root.ui.io.touch.pos) || edit.setFirstEditbox {
			asset._VmDraw_TextSelectTouch(value, valueOrigEdit, OsV2{touchPos, lineY}, lineEnd, editable, font, textH, lineH, margin, marginX)
		}

		this_uid := st.stack //.Hash()
		edit_uid := edit.uid
		active = (edit_uid != nil && edit_uid == this_uid)

		edit.last_edit = value
		if active {
			if lineY == edit.end.Y {
				asset._VmDraw_TextSelectKeys(value, lineY, lineEnd, editable)
			}

			if editable {
				value = asset._VmDraw_TextEditKeys(tabIsChar, font, textH, lineH, margin, marginX) //rewrite 'str' with temp value

				//enter or Tab(key) or outside => save
				isOutside := false
				if touch.start && !st.stack.data.touch_inside {
					uid := edit.uid
					isOutside = (uid != nil && uid == st.stack)
				}
				isEnter := keys.enter
				isEsc := keys.esc
				isTab := !tabIsChar && keys.tab

				if isTab || isEnter || isOutside || isEsc {

					if isEsc {
						//recover
						value = edit.orig
					} //else {
					//save
					//}

					//reset
					edit.uid = nil
					edit.temp = ""
				}
				if isTab {
					edit.tab = true //edit
				}

				//end
				cursorPos = edit.end

				edit.last_edit = value
			}

			//draw selection rectangle
			{
				s := edit.start
				e := edit.end

				if s.Y > e.Y {
					s, e = e, s //swap
				}

				var sx, ex int
				if s.Y != e.Y {
					//multi line
					sx = s.X
					ex = e.X
					if lineY == s.Y {
						ex = utf8.RuneCountInString(value)
					} else if lineY == e.Y {
						sx = 0
					} else if lineY > s.Y && lineY < e.Y {
						sx = 0
						ex = utf8.RuneCountInString(value)
					} else {
						sx = 0
						ex = 0
					}
				} else if lineY == s.Y {
					//one line
					sx = OsMin(s.X, e.X)
					ex = OsMax(s.X, e.X)
				}

				st.buff.AddTextBack(OsV2{sx, ex}, value, coord, font, OsCd_Aprox(OsCd_black(), OsCd_white(), 0.5), textH, align, false, false)
			}
		}
	}

	/*if syntaxtBack != nil {
		for _, it := range syntaxtBack.subs {
			root.ui.PaintTextBack(it, str, coord, font, it.GetColor(), textH, align, false, true)
		}
	}

	if syntaxtUnderline != nil {
		for _, it := range syntaxtUnderline.subs {
			root.ui.PaintTextBack(it, str, coord, font, it.GetColor(), textH, align, true, true)
		}
	}

	if syntaxtLabel != nil {
		for _, it := range syntaxtLabel.subs {
			root.ui.PaintTextTile(str, it, it, coord, font, it.GetColor(), textH, align)
		}
	}*/

	var cds []OsCd
	/*if syntaxtText != nil {
		strN := len(str)
		cds = root.ui.AllocColors(strN, cd)

		for _, it := range syntaxtText.subs {

			cdIt := it.GetColor()
			rng := it
			rng.Sort()
			for j := rng.X; j < strN && j < rng.Y; j++ {
				cds[j] = cdIt
			}
		}
	}*/

	// draw
	st.buff.AddText(value, coord, font, cd, textH, align, cds)

	if cursorPos.X >= 0 {
		//cursor moved
		if !edit.end.Cmp(oldCursorPos) {
			root.ui.SetTextCursorMove()
		}

		var err error
		_ /*cCursorQuad*/, err = st.buff.AddTextCursor(value, coord, font, cd, textH, align, cursorPos.X, root.ui.Cell())
		if err != nil {
			fmt.Println("Error: VmDraw_Text.PaintTextCursor() failed: %w", err)
			return false
		}
	}

	return active
}
