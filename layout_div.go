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
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

type LayoutDiv struct {
	parent *LayoutDiv

	name string

	grid OsV4

	canvas OsV4
	crop   OsV4 //scroll bars NOT included

	data Layout

	enableInput bool

	childs    []*LayoutDiv
	lastChild *LayoutDiv

	gridLock bool //app must set cols/rows first, then draw div(s)
	use      bool //for maintenance

	touchResizeIndex int
	touchResizeCol   bool
}

func (div *LayoutDiv) Use() {
	div.use = true
}

func (div *LayoutDiv) CropWithScroll(ui *Ui) OsV4 {

	ret := div.crop

	if div.data.scrollV.Is() {
		ret.Size.X += div.data.scrollV._GetWidth(ui)
	}

	if div.data.scrollH.Is() {
		ret.Size.Y += div.data.scrollH._GetWidth(ui)
	}
	return ret
	//return OsV4{Start: div.canvas.Start, Size: OsV2{X: div.data.scrollH.screen_height, Y: div.data.scrollV.screen_height}}
}

func (div *LayoutDiv) Print(newLine bool) {
	if div.parent != nil {
		div.parent.Print(false)
	}
	fmt.Printf("[%s : %d, %d, %d, %d]", div.name, div.grid.Start.X, div.grid.Start.Y, div.grid.Size.X, div.grid.Size.Y)
	if newLine {
		fmt.Printf("\n")
	}
}

func (div *LayoutDiv) GetParent(deep int) *LayoutDiv {

	act := div
	for deep > 0 && act != nil {
		act = act.parent
		deep--
	}
	return act
}

func (div *LayoutDiv) Hash() float64 {
	var tmp [8]byte

	h := sha256.New()
	for div != nil {

		binary.LittleEndian.PutUint64(tmp[:], uint64(div.grid.Start.X))
		h.Write(tmp[:])
		binary.LittleEndian.PutUint64(tmp[:], uint64(div.grid.Start.Y))
		h.Write(tmp[:])
		binary.LittleEndian.PutUint64(tmp[:], uint64(div.grid.Size.X))
		h.Write(tmp[:])
		binary.LittleEndian.PutUint64(tmp[:], uint64(div.grid.Size.Y))
		h.Write(tmp[:])

		h.Write([]byte(div.name))

		div = div.parent
	}
	return float64(binary.LittleEndian.Uint64(h.Sum(nil)))
}

func NewLayoutPack(parent *LayoutDiv, name string, grid OsV4, infoLayout *RS_LScroll) *LayoutDiv {

	var div LayoutDiv

	div.name = name
	div.parent = parent
	div.grid = grid

	div.data.Init(div.Hash(), infoLayout)

	div.Use()

	return &div
}

func (div *LayoutDiv) ClearChilds(infoLayout *RS_LScroll) {

	for _, it := range div.childs {
		it.Destroy(infoLayout)
	}
	div.childs = []*LayoutDiv{}
}

func (div *LayoutDiv) Destroy(infoLayout *RS_LScroll) {

	div.ClearChilds(infoLayout)
	div.data.Close(div.Hash(), infoLayout)
}

func (div *LayoutDiv) FindOrCreate(name string, grid OsV4, infoLayout *RS_LScroll) *LayoutDiv {

	//finds
	for _, it := range div.childs {
		if strings.EqualFold(it.name, name) && it.grid.Cmp(grid) {
			div.lastChild = it
			it.Use()
			return it
		}
	}

	// creates
	l := NewLayoutPack(div, name, grid, infoLayout)
	div.childs = append(div.childs, l)
	div.lastChild = l
	return l
}

func (div *LayoutDiv) FindInside(gridPos OsV2) *LayoutDiv {

	var last *LayoutDiv
	for _, it := range div.childs {
		if it.grid.Inside(gridPos) {
			last = it
		}
	}
	return last
}

func (div *LayoutDiv) FindPos(pos OsV2) *LayoutDiv {

	for _, it := range div.childs {
		if it.canvas.Inside(pos) {
			return it
		}
	}
	return nil
}

func (div *LayoutDiv) GetGridMax(minSize OsV2) OsV2 {
	mx := minSize
	for _, it := range div.childs {
		mx = mx.Max(it.grid.End())
	}

	mx = mx.Max(OsV2{div.data.cols.NumIns(), div.data.rows.NumIns()})

	return mx
}

func (div *LayoutDiv) GetLevelSize(winRect OsV4, ui *Ui) OsV4 {

	cell := ui.Cell()
	q := OsV4{OsV2{}, div.GetGridMax(OsV2{1, 1})}

	q.Size = div.data.ConvertMax(cell, q).Size
	q.Start = winRect.Start

	q = q.GetIntersect(winRect)
	return q
}

func (div *LayoutDiv) Maintenance(infoLayout *RS_LScroll) {

	div.use = false

	for i := len(div.childs) - 1; i >= 0; i-- {
		it := div.childs[i]
		if !it.use {
			it.Destroy(infoLayout)

			//remove it
			copy(div.childs[i:], div.childs[i+1:])
			div.childs = div.childs[:len(div.childs)-1]
		} else {
			it.Maintenance(infoLayout)
		}
	}
}

func (div *LayoutDiv) updateGridAndScroll(screen *OsV2, gridMax OsV2, makeSmallerX *bool, makeSmallerY *bool, ui *Ui) bool {

	// update cols/rows
	div.data.UpdateArray(ui.Cell(), *screen, gridMax)

	// get max
	data := div.data.Convert(ui.Cell(), OsV4{OsV2{}, gridMax}).Size

	// make canvas smaller
	hasScrollV := OsTrnBool(*makeSmallerX, data.Y > screen.Y, false)
	hasScrollH := OsTrnBool(*makeSmallerY, data.X > screen.X, false)
	if hasScrollV {
		screen.X -= div.data.scrollV._GetWidth(ui)
		*makeSmallerX = false
	}
	if hasScrollH {
		screen.Y -= div.data.scrollH._GetWidth(ui)
		*makeSmallerY = false
	}

	// save to scroll
	div.data.scrollV.data_height = data.Y
	div.data.scrollV.screen_height = screen.Y

	div.data.scrollH.data_height = data.X
	div.data.scrollH.screen_height = screen.X

	return hasScrollV || hasScrollH
}

func (div *LayoutDiv) UpdateGrid(ui *Ui) {

	makeSmallerX := div.data.scrollV.show
	makeSmallerY := div.data.scrollH.show
	gridMax := div.GetGridMax(OsV2{})

	screen := div.canvas.Size
	for div.updateGridAndScroll(&screen, gridMax, &makeSmallerX, &makeSmallerY, ui) {
	}
}

func (div *LayoutDiv) UpdateCoord(ui *Ui) {

	parent := div.parent
	//backup := laypack.canvas

	if parent != nil {
		div.canvas = parent.data.Convert(ui.Cell(), div.grid)
		div.canvas.Start = parent.canvas.Start.Add(div.canvas.Start)

		// move start by scroll
		div.canvas.Start.Y -= parent.data.scrollV.GetWheel()
		div.canvas.Start.X -= parent.data.scrollH.GetWheel()
	}

	if parent != nil {
		div.crop = div.canvas.GetIntersect(parent.crop)
	}

	// cut 'crop' by scrollbars space
	if div.data.scrollH.Is() {
		div.crop.Size.Y = OsMax(0, div.crop.Size.Y-div.data.scrollH._GetWidth(ui))
	}
	if div.data.scrollV.Is() {
		div.crop.Size.X = OsMax(0, div.crop.Size.X-div.data.scrollV._GetWidth(ui))
	}
}

func (div *LayoutDiv) RenderResizeDraw(layoutScreen OsV4, i int, cd OsCd, vertical bool, buff *PaintBuff) {

	cell := buff.ui.Cell()

	rspace := LayoutArray_resizerSize(cell)
	if vertical {
		layoutScreen.Start.X -= div.data.scrollH.GetWheel()

		layoutScreen.Start.X += div.data.cols.GetResizerPos(i, cell)
		layoutScreen.Size.X = rspace
	} else {
		layoutScreen.Start.Y -= div.data.scrollV.GetWheel()

		layoutScreen.Start.Y += div.data.rows.GetResizerPos(i, cell)
		layoutScreen.Size.Y = rspace
	}

	buff.AddRect(layoutScreen.AddSpace(4), cd, 0)
}

func (div *LayoutDiv) RenderResizeSpliter(root *Root, buff *PaintBuff) {

	enableInput := div.enableInput

	cell := root.ui.Cell()
	tpos := div.GetRelativePos(root.ui.io.touch.pos)

	vHighlight := false
	hHighlight := false
	col := -1
	row := -1
	if enableInput && div.crop.Inside(root.ui.io.touch.pos) {
		col = div.data.cols.IsResizerTouch((tpos.X), cell)
		row = div.data.rows.IsResizerTouch((tpos.Y), cell)

		vHighlight = (col >= 0)
		hHighlight = (row >= 0)

		// start
		if root.ui.io.touch.start && (vHighlight || hHighlight) {
			if vHighlight || hHighlight {
				root.touch.Set(nil, nil, nil, div)
			}

			if vHighlight {
				div.touchResizeIndex = col
				div.touchResizeCol = true
			}
			if hHighlight {
				div.touchResizeIndex = row
				div.touchResizeCol = false
			}
		}

		if root.touch.IsAnyActive() {
			vHighlight = false
			hHighlight = false
			//active = true
		}

		// resize
		if root.touch.IsFnMove(nil, nil, nil, div) {

			r := 1.0
			if div.touchResizeCol {
				col = div.touchResizeIndex
				vHighlight = true

				if div.data.cols.IsLastResizeValid() && int(col) == div.data.cols.NumIns()-2 {
					r = float64(div.canvas.Size.X - tpos.X) // last
				} else {
					r = float64(tpos.X - div.data.cols.GetResizerPos(int(col)-1, cell))
				}

				div.SetResizer(int(col), r, true, root.ui)
			} else {
				row = div.touchResizeIndex
				hHighlight = true

				if div.data.rows.IsLastResizeValid() && int(row) == div.data.rows.NumIns()-2 {
					r = float64(div.canvas.Size.Y - tpos.Y) // last
				} else {
					r = float64(tpos.Y - (div.data.rows.GetResizerPos(int(row)-1, cell)))
				}

				div.SetResizer(int(row), r, false, root.ui)
			}
		}
	}

	// draw all(+active)
	{
		activeCd := root.ui.io.GetThemeCd()
		activeCd.A = 150

		defaultCd := OsCd_Aprox(OsCd_white(), OsCd_black(), 0.3)

		for i := 0; i < div.data.cols.NumIns(); i++ {
			if div.data.cols.GetResizeIndex(i) >= 0 {
				if vHighlight && i == int(col) {
					div.RenderResizeDraw(div.canvas, i, activeCd, true, buff)
				} else {
					div.RenderResizeDraw(div.canvas, i, defaultCd, true, buff)
				}
			}
		}

		for i := 0; i < div.data.rows.NumIns(); i++ {
			if div.data.rows.GetResizeIndex(i) >= 0 {
				if hHighlight && i == int(row) {
					div.RenderResizeDraw(div.canvas, i, activeCd, false, buff)
				} else {
					div.RenderResizeDraw(div.canvas, i, defaultCd, false, buff)
				}
			}
		}
	}

	// cursor
	if enableInput {
		if vHighlight {
			root.ui.PaintCursor("res_col")
		}
		if hHighlight {
			root.ui.PaintCursor("res_row")
		}
	}
}

func (div *LayoutDiv) Save(infoLayout *RS_LScroll) {

	div.data.Save(div.Hash(), infoLayout)

	for _, it := range div.childs {
		it.Save(infoLayout)
	}
}

func (div *LayoutDiv) SetResizer(i int, value float64, isCol bool, ui *Ui) {
	value = math.Max(0.3, value/float64(ui.Cell()))

	var ind int
	var arr *LayoutArray
	if isCol {
		ind = div.data.cols.GetResizeIndex(i)
		arr = &div.data.cols
	} else {
		ind = div.data.rows.GetResizeIndex(i)
		arr = &div.data.rows
	}

	if ind >= 0 {
		arr.items[ind].resize.value = float32(value)
	}
}

func (div *LayoutDiv) GetInputCol(pos int) *LayoutArrayItem {
	return div.data.cols.findOrAdd(pos)
}
func (div *LayoutDiv) GetInputRow(pos int) *LayoutArrayItem {
	return div.data.rows.findOrAdd(pos)
}

func (div *LayoutDiv) GetRelativePos(abs_pos OsV2) OsV2 {
	rpos := abs_pos.Sub(div.canvas.Start)
	rpos.Y += div.data.scrollV.GetWheel()
	rpos.X += div.data.scrollH.GetWheel()
	return rpos
}
