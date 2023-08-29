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
	"math"
	"strings"
)

func themeBack() OsCd {
	return OsCd{220, 220, 220, 255}
}
func themeWhite() OsCd {
	return OsCd{255, 255, 255, 255}
}
func themeBlack() OsCd {
	return OsCd{0, 0, 0, 255}
}
func themeGrey(t float64) OsCd {
	return OsCd{byte(255 * t), byte(255 * t), byte(255 * t), 255}
}

func (asset *Asset) themeCd() OsCd {

	cd := OsCd{90, 180, 180, 255} // ocean
	switch asset.app.root.ui.io.ini.Theme {
	case 1:
		cd = OsCd{200, 100, 80, 255}
	case 2:
		cd = OsCd{130, 170, 210, 255}
	case 3:
		cd = OsCd{130, 180, 130, 255}
	case 4:
		cd = OsCd{160, 160, 160, 255}
	}
	return cd
}

func (asset *Asset) swp_drawButton(backCd OsCd, frontCd OsCd,
	value string, icon string, url string, title string,
	font uint32, alpha float64, alphaNoBack uint32, iconInverseColor uint32,

	margin float64, marginIcon float64, align uint32, ratioH float64,
	enable uint32, highlight uint32, drawBorder uint32) (bool, bool, int64) {

	root := asset.app.root
	st := root.stack

	drawBack := true
	//backCd := OsCd{byte(backCd_r), byte(backCd_g), byte(backCd_b), byte(backCd_a)}
	//frontCd := OsCd{byte(frontCd_r), byte(frontCd_g), byte(frontCd_b), byte(frontCd_a)}

	if alpha == 1 {
		drawBack = false
	} else if alpha == 0.5 {
		drawBack = true
		backCd = OsCd_Aprox(themeBack(), themeBlack(), 0.05)
	}

	if highlight > 0 {
		drawBack = true
		backCd = asset.themeCd()
	}

	if enable == 0 {
		backCd = OsCd_Aprox(themeWhite(), backCd, 0.5)
		frontCd = OsCd_Aprox(themeWhite(), frontCd, 0.8)
	}

	var click, rclick bool
	if enable > 0 {
		active := st.stack.data.touch_active
		inside := st.stack.data.touch_inside
		end := st.stack.data.touch_end
		force := root.ui.io.touch.rm

		if active || inside {
			backCd = OsCd_Aprox(backCd, themeWhite(), 0.5)
			drawBack = true
			if len(icon) > 0 && highlight == 0 {
				drawBack = false
			}

			if alpha > 0 && alphaNoBack > 0 {
				frontCd = asset.themeCd()
				drawBack = false
			}

			asset.paint_cursor("hand")
		}
		if active && inside {
			backCd = themeBack()
			frontCd = asset.themeCd()
		}

		if inside && end {
			click = true
			rclick = force
		}
	}

	if drawBack {
		st.buff.AddRect(asset.getCoord(0, 0, 1, 1, margin, 0, 0), backCd, 0)

		if drawBorder > 0 {
			st.buff.AddRect(asset.getCoord(0, 0, 1, 1, 0, 0, 0), frontCd, asset.getCellWidth(0.03))
		}
	}

	if len(value) > 0 && len(icon) > 0 {

		width := float64(st.stack.canvas.Size.X) / float64(root.ui.Cell())
		height := float64(st.stack.canvas.Size.Y) / float64(root.ui.Cell())

		var w float64
		if height > 0 && width/height > 0 {
			w = 1 / height
		}

		path, err := InitResourcePath(asset.app.root, icon, asset.app.name)
		if err != nil {
			asset.AddLogErr(err)
			return false, false, -1
		}
		st.buff.AddImage(path, iconInverseColor != 0, asset.getCoord(0, 0, w, 1, marginIcon, 0, 0), frontCd, 1, 1, false)

		asset.paint_text(w, 0, 1-w, 1,
			value,
			margin, 0, 0,
			frontCd,
			ratioH, 1,
			0, align, 1,
			0, 0, 0, enable)

	} else if len(value) > 0 {

		asset.paint_text(0, 0, 1, 1,
			value,
			margin, 0, 0,
			frontCd,
			ratioH, 1,
			0, align, 1,
			0, 0, 0, enable)
	} else if len(icon) > 0 {
		path, err := InitResourcePath(asset.app.root, icon, asset.app.name)
		if err != nil {
			asset.AddLogErr(err)
			return false, false, -1
		}

		if enable == 0 {
			frontCd = OsCd_Aprox(themeWhite(), frontCd, 0.5) //boost shades
		}
		st.buff.AddImage(path, iconInverseColor != 0, asset.getCoord(0, 0, 1, 1, marginIcon, 0, 0), frontCd, 1, 1, false)
	}

	if click && len(url) > 0 {
		//SA_DialogStart() warning which open dialog ...
		OsUlit_OpenBrowser(url)
	}

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	} else if len(url) > 0 {
		asset.paint_title(0, 0, 1, 1, url)
	}

	return click, rclick, 1
}

func (asset *Asset) _sa_swp_drawButton(backCd_r, backCd_g, backCd_b, backCd_a uint32,
	frontCd_r, frontCd_g, frontCd_b, frontCd_a uint32,
	valueMem uint64, iconMem uint64, urlMem uint64, titleMem uint64,
	font uint32, alpha float64, alphaNoBack uint32, iconInverseColor uint32,

	margin float64, marginIcon float64, align uint32, ratioH float64,
	enable uint32, highlight uint32, drawBorder uint32,
	outMem uint64) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}
	icon, err := asset.ptrToString(iconMem)
	if asset.AddLogErr(err) {
		return -1
	}
	url, err := asset.ptrToString(urlMem)
	if asset.AddLogErr(err) {
		return -1
	}
	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	click, rclick, ret := asset.swp_drawButton(InitOsCd32(backCd_r, backCd_g, backCd_b, backCd_a),
		InitOsCd32(frontCd_r, frontCd_g, frontCd_b, frontCd_a),
		value, icon, url, title,
		font, alpha, alphaNoBack, iconInverseColor,

		margin, marginIcon, align, ratioH,
		enable, highlight, drawBorder)

	out, err := asset.ptrToBytesDirect(outMem)
	if asset.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(click, 1, 0)))  //click
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(rclick, 1, 0))) //r-click
	return ret
}

func (asset *Asset) swp_drawProgress(value float64, maxValue float64, title string, margin float64, enable uint32) int64 {

	frontCd := asset.themeCd()
	backCd := themeWhite()

	if enable == 0 {
		frontCd = OsCd_Aprox(themeWhite(), frontCd, 0.3)
	}

	w := OsClampFloat(value, 0, maxValue) / maxValue
	asset._sa_paint_rect(0, 0, 1, 1, margin, uint32(backCd.R), uint32(backCd.G), uint32(backCd.B), uint32(backCd.A), 0)
	asset._sa_paint_rect(0, 0, 1, 1, margin, uint32(frontCd.R), uint32(frontCd.G), uint32(frontCd.B), uint32(frontCd.A), 0.03)
	asset._sa_paint_rect(0, 0, w, 1, margin+0.06, uint32(frontCd.R), uint32(frontCd.G), uint32(frontCd.B), uint32(frontCd.A), 0)
	return 1
}

func (asset *Asset) _sa_swp_drawProgress(value float64, maxValue float64, titleMem uint64, margin float64, enable uint32) int64 {

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.swp_drawProgress(value, maxValue, title, margin, enable)
}

func (asset *Asset) swp_drawSlider(value float64, minValue float64, maxValue float64, jumpValue float64, title string, enable uint32) (float64, bool, bool, bool) {

	root := asset.app.root
	st := root.stack

	old_value := value

	frontCd := asset.themeCd()
	backCd := themeGrey(0.75)

	active := st.stack.data.touch_active
	inside := st.stack.data.touch_inside
	end := st.stack.data.touch_end

	cell := float64(asset.app.root.ui.Cell())
	rad := 0.2 / (float64(st.stack.canvas.Size.Y) / cell)
	sp := 0.2 / (float64(st.stack.canvas.Size.X) / cell)

	rpos := root.ui.io.touch.pos.Sub(st.stack.canvas.Start)
	touch_x := float64(rpos.X) / float64(st.stack.canvas.Size.X)

	if enable > 0 {
		if active || inside {
			frontCd = OsCd_Aprox(frontCd, themeWhite(), 0.2)
			backCd = OsCd_Aprox(backCd, themeWhite(), 0.5)
			asset.paint_cursor("hand")
		}

		if active {
			//cut space from touch_x: outer(0,1) => inner(0,1)
			touch_x = OsClampFloat(touch_x, sp, 1-sp)
			touch_x = (touch_x - sp) / (1 - 2*sp)

			frontCd = OsCd_Aprox(frontCd, themeWhite(), 0.2)
			value = minValue + (maxValue-minValue)*touch_x

			t := math.Round((value - minValue) / jumpValue)
			value = minValue + t*jumpValue
			value = OsClampFloat(value, minValue, maxValue)
		}

		//end = props.end
	} else {
		frontCd = OsCd_Aprox(themeWhite(), frontCd, 0.3)
	}

	t := (value - minValue) / (maxValue - minValue)
	//inner(0,1) => outer(0,1)
	t = (t + sp) * (1 - 2*sp)

	width := 0.05
	asset._sa_paint_line(0, 0, 1, 1, 0, sp, 0.5, t, 0.5, uint32(frontCd.R), uint32(frontCd.G), uint32(frontCd.B), uint32(frontCd.A), width)
	asset._sa_paint_line(0, 0, 1, 1, 0, t, 0.5, 1-sp, 0.5, uint32(backCd.R), uint32(backCd.G), uint32(backCd.B), uint32(backCd.A), width)

	asset._sa_paint_circle(0, 0, 1, 1, 0, t, 0.5, rad, uint32(frontCd.R), uint32(frontCd.G), uint32(frontCd.B), uint32(frontCd.A), 0)

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return value, active, (active && old_value != value), end
}

func (asset *Asset) _sa_swp_drawSlider(value float64, minValue float64, maxValue float64, jumpValue float64, titleMem uint64, enable uint32, outMem uint64) float64 {

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	value, active, changed, finished := asset.swp_drawSlider(value, minValue, maxValue, jumpValue, title, enable)

	out, err := asset.ptrToBytesDirect(outMem)
	if asset.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(active, 1, 0)))    //active
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(changed, 1, 0)))   //changed
	binary.LittleEndian.PutUint64(out[16:], uint64(OsTrn(finished, 1, 0))) //finished

	return value
}

func (asset *Asset) swp_drawText(cd_r, cd_g, cd_b, cd_a uint32,
	value string, title string, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, alignV uint32, ratioH float64,
	enable uint32, selection uint32) int64 {

	root := asset.app.root
	st := root.stack

	cd := InitOsCd32(cd_r, cd_g, cd_b, cd_a)
	//origCd := cd

	if align == 1 {
		marginX = 0
	}

	if enable == 0 {
		st.stack.data.touch_enabled = false
		cd = OsCd_Aprox(themeWhite(), cd, 0.3)
	} else {
		if selection > 0 && enable > 0 && (st.stack.data.touch_active || st.stack.data.touch_inside) {
			asset.paint_cursor("ibeam")
		}
	}

	st.stack.data.scrollH.narrow = true
	st.stack.data.scrollV.show = false

	asset._sa_div_col(0, OsMaxFloat(asset.div_get_info("layoutWidth", -1, -1), asset.paint_textWidth(value, font, ratioH, -1)+marginX*4+margin*2))
	asset._sa_div_row(0, asset.div_get_info("layoutHeight", -1, -1))

	asset.div_start(0, 0, 1, 1, "")

	asset.paint_text(0, 0, 1, 1,
		value,
		margin, marginX, marginY,
		cd,
		ratioH, 1,
		0, align, alignV,
		selection, 0, 0, enable)

	asset._sa_div_end()

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return 1
}

func (asset *Asset) _sa_swp_drawText(cd_r, cd_g, cd_b, cd_a uint32,
	valueMem uint64, titleMem uint64, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, alignV uint32, ratioH float64,
	enable uint32, selection uint32) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.swp_drawText(cd_r, cd_g, cd_b, cd_a,
		value, title, font,
		margin, marginX, marginY, align, alignV, ratioH,
		enable, selection)
}

func (asset *Asset) swp_getEditValue() string {
	return asset.app.root.ui.io.edit.last_edit
}

func (asset *Asset) _sa_swp_getEditValue(outMem uint64) int64 {
	err := asset.stringToPtr(asset.swp_getEditValue(), outMem)
	if !asset.AddLogErr(err) {
		return -1
	}
	return 1
}

func (asset *Asset) swp_drawEdit(cd_r, cd_g, cd_b, cd_a uint32,
	valueIn string, title string, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, alignV uint32, ratioH float64,
	enable uint32) (string, bool, bool, bool) {

	root := asset.app.root
	div := root.stack.stack

	cd := InitOsCd32(cd_r, cd_g, cd_b, cd_a)
	if align == 1 {
		marginX = 0
	}

	div.data.scrollH.narrow = true
	div.data.scrollV.show = false

	edit := &root.ui.io.edit

	inDiv := root.stack.stack.FindOrCreate("", InitOsQuad(0, 0, 1, 1), &root.infoLayout)
	this_uid := inDiv //.Hash()
	edit_uid := edit.uid
	active := (edit_uid != nil && edit_uid == this_uid)

	var value string
	if active {
		value = edit.temp
	} else {
		value = valueIn
	}

	if enable == 0 {
		cd = OsCd_Aprox(themeWhite(), cd, 0.3)
	} else if div.data.touch_active || div.data.touch_inside {
		asset.paint_cursor("ibeam")
	}
	inDiv.data.touch_enabled = (enable != 0)

	asset._sa_div_col(0, OsMaxFloat(asset.div_get_info("layoutWidth", -1, -1), asset.paint_textWidth(value, font, ratioH, -1)+marginX*4+margin*2))
	asset._sa_div_row(0, asset.div_get_info("layoutHeight", -1, -1))

	asset.div_start(0, 0, 1, 1, "")

	asset.paint_text(0, 0, 1, 1,
		value,
		margin, marginX, marginY,
		cd,
		ratioH, 1,
		0, align, alignV,
		1, 1, 1, enable)

	asset._sa_div_end()

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return edit.last_edit, active, (active && value != edit.last_edit), (active && this_uid != edit.uid)
}

func (asset *Asset) _sa_swp_drawEdit(cd_r, cd_g, cd_b, cd_a uint32,
	valueMem uint64, titleMem uint64, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, alignV uint32, ratioH float64,
	enable uint32,
	outMem uint64) int64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	last_edit, active, changed, finished := asset.swp_drawEdit(cd_r, cd_g, cd_b, cd_a,
		value, title, font,
		margin, marginX, marginY, align, alignV, ratioH, enable)

	out, err := asset.ptrToBytesDirect(outMem)
	if asset.AddLogErr(err) {
		return -1
	}
	binary.LittleEndian.PutUint64(out[0:], uint64(OsTrn(active, 1, 0)))    //active
	binary.LittleEndian.PutUint64(out[8:], uint64(OsTrn(changed, 1, 0)))   //changed
	binary.LittleEndian.PutUint64(out[16:], uint64(OsTrn(finished, 1, 0))) //finished
	binary.LittleEndian.PutUint64(out[24:], uint64(len(last_edit)))        //size
	return 1
}

func (asset *Asset) swp_drawCombo(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, optionsIn string, title string, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, ratioH float64,
	enable uint32) int64 {

	cd := InitOsCd32(cd_r, cd_g, cd_b, cd_a)
	origCd := cd
	if enable == 0 {
		cd = OsCd_Aprox(OsCd_white(), cd, 0.3)
	}

	root := asset.app.root
	div := root.stack.stack

	options := strings.Split(optionsIn, "|")
	var val string
	if value >= uint64(len(options)) {
		val = ""
	} else {
		val = options[value]
	}

	w := 0.6 / (float64(div.canvas.Size.X) / float64(asset.app.root.ui.Cell()))

	//text
	asset.paint_text(0, 0, 1-w, 1,
		val,
		margin, marginX, marginY,
		origCd,
		ratioH, 1,
		0, align, 1,
		0, 0, 0, enable)

	//arrow
	asset.paint_text(1-w, 0, w, 1,
		"▼",
		margin, 0, 0,
		origCd,
		ratioH, 1,
		0, align, 1,
		0, 0, 0, enable)

	//border
	asset.paint_rect(0, 0, 1, 1, 0, cd, 0.03)

	if enable > 0 {
		//cursor
		if div.data.touch_active || div.data.touch_inside {
			asset.paint_cursor("hand")
		}
	}

	//dialog
	if asset.div_dialogStart("combo", 1, div.data.touch_end && enable > 0) > 0 {
		//compute minimum dialog width
		mx := 0
		for _, opt := range options {
			mx = OsMax(mx, len(opt))
		}
		asset._sa_div_colMax(0, OsMaxFloat(5, ratioH*float64(mx)))

		for i, opt := range options {
			asset.div_start(0, uint64(i), 1, 1, "")
			click, _, ret := asset.swp_drawButton(OsCd_black(), cd, opt, "", "", "", 0, 1, 1, 0, margin, 0, 0, ratioH, 1, uint32(OsTrn(value == uint64(i), 1, 0)), 0)
			if ret > 0 && click {
				value = uint64(i)
				asset._sa_div_dialogClose()
				break
			}

			asset._sa_div_end()
		}

		asset._sa_div_dialogEnd()
	}

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return int64(value)
}

func (asset *Asset) _sa_swp_drawCombo(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, optionsMem uint64, titleMem uint64, font uint32,
	margin float64, marginX float64, marginY float64, align uint32, ratioH float64,
	enable uint32) int64 {

	options, err := asset.ptrToString(optionsMem)
	if asset.AddLogErr(err) {
		return -1
	}
	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}
	return asset.swp_drawCombo(cd_r, cd_g, cd_b, cd_a,
		value, options, title, font,
		margin, marginX, marginY, align, ratioH, enable)
}

func (asset *Asset) swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a uint32,
	value uint64, description string, title string, enable uint32) int64 {

	root := asset.app.root
	st := root.stack

	cd := InitOsCd32(cd_r, cd_g, cd_b, cd_a)
	origCd := cd

	if enable > 0 {
		active := st.stack.data.touch_active
		inside := st.stack.data.touch_inside
		end := st.stack.data.touch_end

		if active || inside {
			cd = OsCd_Aprox(cd, OsCd_white(), 0.3)
			asset.paint_cursor("hand")
		}

		if inside && end {
			value = uint64(OsTrn(value > 0, 0, 1))
		}

	} else {
		cd = OsCd_Aprox(OsCd_white(), cd, 0.3)
	}

	width := float64(st.stack.canvas.Size.X) / float64(root.ui.Cell())
	height := float64(st.stack.canvas.Size.Y) / float64(root.ui.Cell())
	w := 1 / (width / height)

	if value > 0 {
		asset.paint_rect(0, 0, w, 1, 0.22, cd, 0)
		asset._sa_paint_line(0, 0, w, 1, 0.33, 1.0/3, 0.9, 0.05, 2.0/3, 255, 255, 255, 255, 0.05)
		asset._sa_paint_line(0, 0, w, 1, 0.33, 1.0/3, 0.9, 0.95, 1.0/4, 255, 255, 255, 255, 0.05)
	} else {
		asset.paint_rect(0, 0, w, 1, 0.22, cd, 0.03)
	}

	asset.paint_text(w*0.8, 0, 1-w*0.8, 1, description, 0, 0.1, 0, origCd, 0.35, 1, 0, 0, 1, 0, 0, 0, enable)

	if len(title) > 0 {
		asset.paint_title(0, 0, 1, 1, title)
	}

	return int64(value)
}

func (asset *Asset) _sa_swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a uint32, value uint64, descriptionMem uint64, titleMem uint64, enable uint32) int64 {

	description, err := asset.ptrToString(descriptionMem)
	if asset.AddLogErr(err) {
		return -1
	}

	title, err := asset.ptrToString(titleMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.swp_drawCheckbox(cd_r, cd_g, cd_b, cd_a, value, description, title, enable)
}

func (asset *Asset) paint_textWidth(value string, fontId uint32, ratioH float64, cursorPos int64) float64 {

	root := asset.app.root

	textH := asset.getCellWidth(ratioH)
	font := root.fonts.Get(int(fontId))
	cell := float64(asset.app.root.ui.Cell())
	if cursorPos < 0 {

		size, err := font.GetTextSize(value, textH, 0)
		if err == nil {
			return float64(size.X) / cell // pixels for the whole string
		}

	} else {
		px, err := font.GetPxPos(value, textH, int(cursorPos))
		if err == nil {
			return float64(px) / cell // pixels to cursor
		}
	}
	return -1
}

func (asset *Asset) _sa_paint_textWidth(valueMem uint64, fontId uint32, ratioH float64, cursorPos int64) float64 {

	value, err := asset.ptrToString(valueMem)
	if asset.AddLogErr(err) {
		return -1
	}

	return asset.paint_textWidth(value, fontId, ratioH, cursorPos)
}