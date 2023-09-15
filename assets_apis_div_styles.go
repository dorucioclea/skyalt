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
	"encoding/json"
	"fmt"
)

type DivStyle struct {
	Margin_top, Margin_bottom, Margin_left, Margin_right     float64 //from cell
	Border_top, Border_bottom, Border_left, Border_right     float64 //from cell
	Padding_top, Padding_bottom, Padding_left, Padding_right float64 //from cell

	Margin_top_color, Margin_bottom_color, Margin_left_color, Margin_right_color     OsCd
	Border_top_color, Border_bottom_color, Border_left_color, Border_right_color     OsCd
	Padding_top_color, Padding_bottom_color, Padding_left_color, Padding_right_color OsCd
	Content_color                                                                    OsCd

	Image_fill                 bool
	Image_alignV, Image_alignH int

	Font_color               OsCd
	Font_path                string
	Font_height              float64 //from cell
	Font_alignV, Font_alignH int

	Cursor string

	//radius ...
	//shadow ...
	//transition_sec(blend between states) ...
}

func (st *DivStyle) Margin(v float64) {
	st.Margin_top = v
	st.Margin_bottom = v
	st.Margin_left = v
	st.Margin_right = v
}
func (st *DivStyle) Border(v float64) {
	st.Border_top = v
	st.Border_bottom = v
	st.Border_left = v
	st.Border_right = v
}
func (st *DivStyle) BorderCd(v OsCd) {
	st.Border_top_color = v
	st.Border_bottom_color = v
	st.Border_left_color = v
	st.Border_right_color = v
}

func _paintBorder(out OsV4, top, bottom, left, right float64, topCd, bottomCd, leftCd, rightCd OsCd, asset *Asset) OsV4 {

	stt := asset.app.root.levels.GetStack()

	in := out.Inner(asset.getCellWidth(top), asset.getCellWidth(bottom), asset.getCellWidth(left), asset.getCellWidth(right))

	if topCd.A > 0 {
		q := OsV4{Start: out.Start, Size: OsV2{out.Size.X, asset.getCellWidth(top)}}
		stt.buff.AddRect(q, topCd, 0)
	}
	if bottomCd.A > 0 {
		q := OsV4{Start: OsV2{out.Start.X, in.Start.Y + in.Size.Y}, Size: OsV2{out.Size.X, asset.getCellWidth(bottom)}}
		stt.buff.AddRect(q, bottomCd, 0)
	}

	if leftCd.A > 0 {
		q := OsV4{Start: out.Start, Size: OsV2{asset.getCellWidth(left), out.Size.Y}}
		stt.buff.AddRect(q, leftCd, 0)
	}

	if rightCd.A > 0 {
		q := OsV4{Start: OsV2{in.Start.X + in.Size.X, out.Start.Y}, Size: OsV2{asset.getCellWidth(right), out.Size.Y}}
		stt.buff.AddRect(q, rightCd, 0)
	}

	return in
}

func DivStyle_getCoord(coord OsV4, x, y, w, h float64) OsV4 {

	return InitOsQuad(
		coord.Start.X+int(float64(coord.Size.X)*x),
		coord.Start.Y+int(float64(coord.Size.Y)*y),
		int(float64(coord.Size.X)*w),
		int(float64(coord.Size.Y)*h))
}

func (st *DivStyle) Paint(coord OsV4, text string, image_path string, image_margin float64, inside bool, asset *Asset) OsV4 {

	stt := asset.app.root.levels.GetStack()
	if stt.stack == nil || stt.stack.crop.IsZero() {
		return OsV4{}
	}

	border := _paintBorder(coord, st.Margin_top, st.Margin_bottom, st.Margin_left, st.Margin_right, st.Margin_top_color, st.Margin_bottom_color, st.Margin_left_color, st.Margin_right_color, asset)
	padding := _paintBorder(border, st.Border_top, st.Border_bottom, st.Border_left, st.Border_right, st.Border_top_color, st.Border_bottom_color, st.Border_left_color, st.Border_right_color, asset)
	content := _paintBorder(padding, st.Padding_top, st.Padding_bottom, st.Padding_left, st.Padding_right, st.Padding_top_color, st.Padding_bottom_color, st.Padding_left_color, st.Padding_right_color, asset)

	if st.Content_color.A > 0 {
		stt.buff.AddRect(content, st.Content_color, 0)
	}

	coordImg := content
	coordText := content

	if len(image_path) > 0 && len(text) > 0 {

		w := float64(asset.app.root.ui.Cell()) / float64(stt.stack.canvas.Size.X)

		switch st.Image_alignH {
		case 0: //left
			coordImg = DivStyle_getCoord(content, 0, 0, w, 1)
			coordText = DivStyle_getCoord(content, w, 0, 1-w, 1)
		case 1: //center
			coordImg = coord
			coordText = coord
		default: //right
			coordImg = DivStyle_getCoord(content, 1-w, 0, w, 1)
			coordText = DivStyle_getCoord(content, 0, 0, 1-w, 1)
		}
	}

	if len(image_path) > 0 {
		path, err := InitResourcePath(asset.app.root, image_path, asset.app.name)
		if err != nil {
			asset.AddLogErr(err)
		} else {
			coordImg = coordImg.Inner(asset.getCellWidth(image_margin), asset.getCellWidth(image_margin), asset.getCellWidth(image_margin), asset.getCellWidth(image_margin))

			stt.buff.AddImage(path, false, coordImg, st.Font_color, st.Image_alignV, st.Image_alignH, st.Image_fill)
		}
	}

	if len(text) > 0 {
		font := asset.app.root.fonts.Get(st.Font_path)
		if font == nil {
			asset.AddLogErr(fmt.Errorf("Font(%s) not found", st.Font_path))
			font = asset.app.root.fonts.Get(SKYALT_FONT_0) //load default
		}
		if font != nil {
			stt.buff.AddText(text, coordText, font, st.Font_color, asset.getCellWidth(st.Font_height), OsV2{st.Font_alignH, st.Font_alignV}, nil)
		}
	}

	if inside && len(st.Cursor) > 0 {
		asset.paint_cursor(st.Cursor)
	}

	return content
}

type SwpStyle struct {
	Main        DivStyle
	Hover       DivStyle
	Touch_hover DivStyle
	Touch_out   DivStyle
	Disable     DivStyle
}

func (b *SwpStyle) FontH(v float64) *SwpStyle {
	b.Main.Font_height = v
	b.Hover.Font_height = v
	b.Touch_hover.Font_height = v
	b.Touch_out.Font_height = v
	b.Disable.Font_height = v
	return b
}

func (b *SwpStyle) FontAlignH(v int) *SwpStyle {
	b.Main.Font_alignH = v
	b.Hover.Font_alignH = v
	b.Touch_hover.Font_alignH = v
	b.Touch_out.Font_alignH = v
	b.Disable.Font_alignH = v
	return b
}
func (b *SwpStyle) FontAlignV(v int) *SwpStyle {
	b.Main.Font_alignV = v
	b.Hover.Font_alignV = v
	b.Touch_hover.Font_alignV = v
	b.Touch_out.Font_alignV = v
	b.Disable.Font_alignV = v
	return b
}

func (b *SwpStyle) Margin(v float64) *SwpStyle {
	b.Main.Margin(v)
	b.Hover.Margin(v)
	b.Touch_hover.Margin(v)
	b.Touch_out.Margin(v)
	b.Disable.Margin(v)
	return b
}

func (b *SwpStyle) Border(v float64) *SwpStyle {
	b.Main.Border(v)
	b.Hover.Border(v)
	b.Touch_hover.Border(v)
	b.Touch_out.Border(v)
	b.Disable.Border(v)
	return b
}
func (b *SwpStyle) BorderCd(v OsCd) *SwpStyle {
	b.Main.BorderCd(v)
	b.Hover.BorderCd(v)
	b.Touch_hover.BorderCd(v)
	b.Touch_out.BorderCd(v)
	b.Disable.BorderCd(v)
	return b
}

func (style *SwpStyle) Paint(coord OsV4, text string, image_path string, image_margin float64, enable bool, asset *Asset) (bool, bool) {

	st := asset.app.root.levels.GetStack()

	var click, rclick bool
	if enable {
		active := st.stack.data.touch_active
		inside := st.stack.data.touch_inside
		end := st.stack.data.touch_end
		force := asset.app.root.ui.io.touch.rm

		if active {
			if inside {
				style.Touch_hover.Paint(coord, text, image_path, image_margin, inside, asset)
			} else {
				style.Touch_out.Paint(coord, text, image_path, image_margin, inside, asset)
			}
		} else {
			if inside {
				style.Hover.Paint(coord, text, image_path, image_margin, inside, asset)
			} else {
				style.Main.Paint(coord, text, image_path, image_margin, inside, asset)
			}
		}

		if inside && end {
			click = true
			rclick = force
		}
	} else {
		style.Disable.Paint(coord, text, image_path, image_margin, false, asset)
	}

	return click, rclick
}

type DivDefaultStyles struct {
	Button             SwpStyle
	ButtonLight        SwpStyle
	ButtonAlpha        SwpStyle
	ButtonMenu         SwpStyle
	ButtonMenuSelected SwpStyle

	ButtonBig      SwpStyle
	ButtonLightBig SwpStyle
	ButtonAlphaBig SwpStyle
	ButtonMenuBig  SwpStyle

	ButtonAlphaBorder SwpStyle

	ButtonLogo SwpStyle

	ButtonDanger     SwpStyle
	ButtonDangerMenu SwpStyle
}

func DivStyles_getDefaults(asset *Asset) DivDefaultStyles {
	stls := DivDefaultStyles{}

	{
		b := &stls.Button.Main
		b.Cursor = "hand"
		b.Content_color = asset.themeCd()
		b.Font_color = themeBlack()
		b.Image_alignV = 1
		b.Image_alignH = 0
		b.Font_path = SKYALT_FONT_0
		b.Font_alignV = 1
		b.Font_alignH = 1
		b.Font_height = 0.35
		b.Margin(0.06)

		//copy .main to others
		stls.Button.Hover = *b
		stls.Button.Touch_hover = *b
		stls.Button.Touch_out = *b
		stls.Button.Disable = *b

		stls.Button.Hover.Content_color = OsCd_Aprox(stls.Button.Main.Content_color, themeWhite(), 0.5)
		stls.Button.Touch_out.Content_color = stls.Button.Hover.Content_color

		stls.Button.Touch_hover.Content_color = themeBack()
		stls.Button.Touch_hover.Font_color = asset.themeCd()

		stls.Button.Disable.Font_color = OsCd_Aprox(stls.Button.Main.Font_color, themeWhite(), 0.35)
		stls.Button.Disable.Content_color = OsCd_Aprox(stls.Button.Main.Content_color, themeWhite(), 0.7)
	}

	{
		stls.ButtonLight = stls.Button
		a := byte(127)
		stls.ButtonLight.Main.Content_color.A = a
		stls.ButtonLight.Hover.Content_color.A = a
		stls.ButtonLight.Touch_hover.Content_color.A = a
		stls.ButtonLight.Touch_out.Content_color.A = a
		stls.ButtonLight.Disable.Content_color.A = a
		stls.ButtonLight.Disable.Font_color.A = a
	}

	{
		stls.ButtonAlpha = stls.Button
		stls.ButtonAlpha.Main.Content_color = OsCd{}
		stls.ButtonAlpha.Hover.Content_color = OsCd_Aprox(asset.themeCd(), themeWhite(), 0.7)
		stls.ButtonAlpha.Touch_out.Content_color = OsCd{}
		stls.ButtonAlpha.Disable.Content_color = OsCd{}
		stls.ButtonAlpha.Disable.Font_color = OsCd_Aprox(asset.themeCd(), themeWhite(), 0.7)
	}

	{
		stls.ButtonMenu = stls.ButtonAlpha
		stls.ButtonMenu.FontAlignH(0)
	}

	{
		stls.ButtonMenuSelected = stls.Button
		stls.ButtonMenuSelected.FontAlignH(0)
	}

	{
		stls.ButtonBig = stls.Button
		stls.ButtonLightBig = stls.ButtonLight
		stls.ButtonAlphaBig = stls.ButtonAlpha
		stls.ButtonMenuBig = stls.ButtonMenu
		stls.ButtonBig.FontH(0.45)
		stls.ButtonLightBig.FontH(0.45)
		stls.ButtonAlphaBig.FontH(0.45)
		stls.ButtonMenuBig.FontH(0.45)
	}

	{
		stls.ButtonAlphaBorder = stls.ButtonAlpha
		//stls.ButtonAlphaBorder.Margin(0.1)
		stls.ButtonAlphaBorder.Border(0.03)
		stls.ButtonAlphaBorder.BorderCd(asset.themeCd())
	}

	{
		stls.ButtonLogo = stls.ButtonAlpha
		stls.ButtonLogo.Hover.Content_color.A = 0
		stls.ButtonLogo.Touch_hover.Content_color.A = 0
		stls.ButtonLogo.Touch_out.Content_color.A = 0 //refactor ...

		stls.ButtonLogo.Hover.Font_color = OsCd_Aprox(stls.ButtonLogo.Main.Font_color, themeWhite(), 0.5)
		stls.ButtonLogo.Touch_hover.Font_color = asset.themeCd()
		stls.ButtonLogo.Touch_out.Font_color = stls.ButtonLogo.Hover.Font_color
	}

	{
		stls.ButtonDanger = stls.Button
		stls.ButtonDanger.Main.Content_color = themeWarning()
		stls.ButtonDanger.Hover.Content_color = themeWarning()
		stls.ButtonDanger.Touch_hover.Content_color = themeWarning()
		stls.ButtonDanger.Touch_out.Content_color = themeWarning()
		stls.ButtonDanger.Disable.Content_color = OsCd_Aprox(stls.ButtonDanger.Main.Content_color, themeWhite(), 0.5)
	}

	{
		stls.ButtonDangerMenu = stls.ButtonDanger
		stls.ButtonDangerMenu.FontAlignH(0)
	}
	return stls
}

type DivStyles struct {
	styles []*SwpStyle

	buttonMenu uint32
}

func NewDivStyles(asset *Asset) *DivStyles {
	var stls DivStyles

	stls.Add(&SwpStyle{}) //empty id=0

	defs := DivStyles_getDefaults(asset)
	stls.buttonMenu = uint32(stls.Add(&defs.ButtonMenu))

	return &stls
}

func (stls *DivStyles) Get(i uint32) *SwpStyle {
	if i > 0 && int(i) < len(stls.styles) {
		return stls.styles[i]
	}
	return nil
}

func (stls *DivStyles) Add(st *SwpStyle) int {
	stls.styles = append(stls.styles, st)
	return len(stls.styles) - 1
}

func (stls *DivStyles) AddJs(js []byte) (int, error) {

	var div SwpStyle
	err := json.Unmarshal(js, &div)
	if err != nil {
		return -1, fmt.Errorf("Unmarshal() failed: %w", err)
	}

	return stls.Add(&div), nil
}
