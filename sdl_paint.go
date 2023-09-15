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

	"github.com/veandco/go-sdl2/gfx"
)

const (
	PaintCrop   byte = 0
	PaintRect   byte = 1
	PaintLine   byte = 2
	PaintCircle byte = 3
	PaintPoly   byte = 4
	PaintImage  byte = 5
	PaintText   byte = 6
)

type PaintItem struct {
	tp    byte
	coord OsV4

	cd    OsCd
	thick int

	start OsV2
	end   OsV2

	x []int16
	y []int16

	path        ResourcePath
	inverserRGB bool

	text  string
	font  *Font
	h     int
	align OsV2
	cds   []OsCd
}

func (pnt *PaintItem) Crop(ui *Ui) {
	err := ui.render.SetClipRect(pnt.coord.GetSDLRect())
	if err != nil {
		return
	}
}

func (pnt *PaintItem) Rect(ui *Ui) {
	start := pnt.coord.Start
	end := pnt.coord.End()
	if pnt.thick == 0 {
		_Ui_boxSE(ui.render, start, end, pnt.cd)
	} else {
		_Ui_boxSE_border(ui.render, start, end, pnt.cd, pnt.thick)
	}
}

func (pnt *PaintItem) Line(ui *Ui) {
	v := pnt.end.Sub(pnt.start)
	if !v.IsZero() {
		_Ui_line(ui.render, pnt.start, pnt.end, pnt.thick, pnt.cd)
	}
}

func (pnt *PaintItem) Circle(ui *Ui) {
	p := pnt.coord.Middle()
	if pnt.thick == 0 {
		gfx.FilledEllipseRGBA(ui.render, int32(p.X), int32(p.Y), int32(pnt.coord.Size.X/2), int32(pnt.coord.Size.Y/2), pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
		gfx.AAEllipseRGBA(ui.render, int32(p.X), int32(p.Y), int32(pnt.coord.Size.X/2), int32(pnt.coord.Size.Y/2), pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	} else {
		gfx.AAEllipseRGBA(ui.render, int32(p.X), int32(p.Y), int32(pnt.coord.Size.X/2), int32(pnt.coord.Size.Y/2), pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	}
}

func (pnt *PaintItem) AddPoly(pos OsV2) {
	pnt.x = append(pnt.x, int16(pos.X))
	pnt.y = append(pnt.y, int16(pos.Y))
}

func (pnt *PaintItem) Poly(ui *Ui) {

	if pnt.thick == 0 {
		gfx.FilledPolygonRGBA(ui.render, ui.poly.x, ui.poly.y, pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	} else {
		gfx.AAPolygonRGBA(ui.render, ui.poly.x, ui.poly.y, pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	}
}

func PaintImage_load(path ResourcePath, inverserRGB bool, ui *Ui) (*Image, error) {

	var img *Image
	for _, it := range ui.images {
		if it.path.Cmp(&path) && it.inverserRGB == inverserRGB {
			img = it
			break
		}
	}

	if img == nil {
		var err error
		img, err = NewImage(path, inverserRGB, ui.render)
		if err != nil {
			return nil, fmt.Errorf("NewImage() failed: %w", err)
		}

		if img != nil {
			ui.images = append(ui.images, img)
		}
	}

	return img, nil
}
func (pnt *PaintItem) Image(ui *Ui) error {

	img, err := PaintImage_load(pnt.path, pnt.inverserRGB, ui)
	if err != nil {
		return fmt.Errorf("PaintImage_load() failed: %w", err)
	}

	if img != nil {
		err := img.Draw(pnt.coord, pnt.cd, ui.render)
		if err != nil {
			return fmt.Errorf("Draw() failed: %w", err)
		}
	}
	return nil
}

func (pnt *PaintItem) Text(ui *Ui) {

	err := pnt.font.Print(pnt.text, pnt.h, pnt.coord, pnt.align, pnt.cd, pnt.cds, ui.render)
	if err != nil {
		fmt.Printf("Print() failed: %v\n", err)
		return
	}
}

type PaintBuff struct {
	ui    *Ui
	items []PaintItem

	lastCrop OsV4
}

func NewPaintBuff(ui *Ui) *PaintBuff {
	var b PaintBuff
	b.ui = ui
	return &b
}

func (b *PaintBuff) Reset(crop OsV4) {
	b.lastCrop = crop
	b.items = b.items[:0]
}

func (b *PaintBuff) AddCrop(coord OsV4) OsV4 {

	b.items = append(b.items, PaintItem{tp: PaintCrop, coord: coord})

	old := b.lastCrop
	b.lastCrop = coord
	return old
}

func (b *PaintBuff) AddRect(coord OsV4, cd OsCd, thick int) {
	b.items = append(b.items, PaintItem{tp: PaintRect, coord: coord, cd: cd, thick: thick})
}

func (b *PaintBuff) AddLine(start OsV2, end OsV2, cd OsCd, thick int) {
	b.items = append(b.items, PaintItem{tp: PaintLine, start: start, end: end, cd: cd, thick: thick})
}

func (b *PaintBuff) AddCircle(coord OsV4, cd OsCd, thick int) {
	b.items = append(b.items, PaintItem{tp: PaintCircle, coord: coord, cd: cd, thick: thick})
}

func (b *PaintBuff) AddImage(path ResourcePath, inverserRGB bool, coord OsV4, cd OsCd, alignV int, alignH int, fill bool) {

	img, err := PaintImage_load(path, inverserRGB, b.ui)
	if err != nil {
		b.AddText(path.GetString()+" has error", coord, path.root.fonts.Get(SKYALT_FONT_0), OsCd_error(), path.root.ui.io.GetDPI()/8, OsV2{1, 1}, nil)
		return
	}
	if img == nil {
		return //image is empty
	}

	var q OsV4
	if !fill {
		rect_size := OsV2_InRatio(coord.Size, img.origSize)
		q = OsV4_center(coord, rect_size)
	} else {
		q.Start = coord.Start
		q.Size = OsV2_OutRatio(coord.Size, img.origSize)
	}

	if alignH == 0 {
		q.Start.X = coord.Start.X
	} else if alignH == 1 {
		q.Start.X = OsV4_centerFull(coord, q.Size).Start.X
	} else if alignH == 2 {
		q.Start.X = coord.End().X - q.Size.X
	}

	if alignV == 0 {
		q.Start.Y = coord.Start.Y
	} else if alignV == 1 {
		q.Start.Y = OsV4_centerFull(coord, q.Size).Start.Y
	} else if alignV == 2 {
		q.Start.Y = coord.End().Y - q.Size.Y
	}

	imgRectBackup := b.AddCrop(b.lastCrop.GetIntersect(coord))

	b.items = append(b.items, PaintItem{tp: PaintImage, path: path, inverserRGB: inverserRGB, cd: cd, coord: q})

	b.AddCrop(imgRectBackup)
}

func (b *PaintBuff) AddText(text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, cds []OsCd) {
	b.items = append(b.items, PaintItem{tp: PaintText, text: text, coord: coord, font: font, cd: cd, h: h, align: align, cds: cds})
}

func (b *PaintBuff) AddTextBack(rangee OsV2, text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, underline bool, addSpaceY bool) error {

	if rangee.X == rangee.Y {
		return nil
	}

	start, err := font.Start(text, h, coord, align, nil)
	if err != nil {
		return fmt.Errorf("Start() failed: %w", err)
	}

	var rng OsV2
	rng.X, err = font.GetPxPos(text, h, rangee.X)
	if err != nil {
		return fmt.Errorf("GetPxPos(1) failed: %w", err)
	}
	rng.Y, err = font.GetPxPos(text, h, rangee.Y)
	if err != nil {
		return fmt.Errorf("GetPxPos(2) failed: %w", err)
	}
	rng.Sort()

	if rng.X != rng.Y {
		if underline {
			Y := coord.Start.Y + coord.Size.Y
			b.AddRect(OsV4{Start: OsV2{start.X + rng.X, Y - 2}, Size: OsV2{rng.Y, 2}}, cd, 0)
		} else {
			c := InitOsQuad(start.X+rng.X, coord.Start.Y, rng.Y-rng.X, coord.Size.Y)
			if addSpaceY {
				c = c.AddSpaceY((coord.Size.Y - h) / 4) //smaller height
			}
			b.AddRect(c, cd, 0)
		}
	}
	return nil
}

func (b *PaintBuff) AddTextCursor(text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, cursorPos int, cell int) (OsV4, error) {

	b.ui.cursorEdit = true
	cd.A = b.ui.cursorCdA

	start, err := font.Start(text, h, coord, align, nil)
	if err != nil {
		return OsV4{}, fmt.Errorf("TextCursor().Start() failed: %w", err)
	}

	ex, err := font.GetPxPos(text, h, cursorPos)
	if err != nil {
		return OsV4{}, fmt.Errorf("TextCursor().GetPxPos() failed: %w", err)
	}

	cursorQuad := InitOsQuad(start.X+ex, coord.Start.Y, OsMax(1, cell/15), coord.Size.Y)
	cursorQuad = cursorQuad.AddSpaceY((coord.Size.Y - h) / 4) //smaller height

	b.AddRect(cursorQuad, cd, 0)

	return cursorQuad, nil
}

func (b *PaintBuff) Draw() {

	ui := b.ui

	for _, it := range b.items {
		switch it.tp {
		case PaintCrop:
			it.Crop(ui)
		case PaintRect:
			it.Rect(ui)
		case PaintLine:
			it.Line(ui)
		case PaintCircle:
			it.Circle(ui)
		case PaintPoly:
			it.Poly(ui)
		case PaintImage:
			it.Image(ui)
		case PaintText:
			it.Text(ui)
		}
	}
}

/*type PaintCrop struct {
	coord OsV4
}

func (pnt *PaintCrop) Draw(ui *Ui) {
	err := ui.render.SetClipRect(pnt.coord.GetSDLRect())
	if err != nil {
		return
	}
}

type PaintRect struct {
	coord OsV4
	cd    OsCd
	thick int
}

func (pnt *PaintRect) Draw(ui *Ui) {
	start := pnt.coord.Start
	end := pnt.coord.End()
	if pnt.thick == 0 {
		_Ui_boxSE(ui.render, start, end, pnt.cd)
	} else {
		_Ui_boxSE_border(ui.render, start, end, pnt.cd, pnt.thick)
	}
}

type PaintLine struct {
	start OsV2
	end   OsV2
	cd    OsCd
	thick int
}

func (pnt *PaintLine) Draw(ui *Ui) {
	v := pnt.end.Sub(pnt.start)
	if !v.IsZero() {
		_Ui_line(ui.render, pnt.start, pnt.end, pnt.thick, pnt.cd)
	}
}

type PaintCircle struct {
	coord OsV4
	cd    OsCd
	thick int
}

func (pnt *PaintCircle) Draw(ui *Ui) {
	p := pnt.coord.Middle()
	if pnt.thick == 0 {
		gfx.FilledEllipseRGBA(ui.render, int32(p.X), int32(p.Y), int32(pnt.coord.Size.X/2), int32(pnt.coord.Size.Y/2), pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
		gfx.AAEllipseRGBA(ui.render, int32(p.X), int32(p.Y), int32(pnt.coord.Size.X/2), int32(pnt.coord.Size.Y/2), pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	} else {
		gfx.AAEllipseRGBA(ui.render, int32(p.X), int32(p.Y), int32(pnt.coord.Size.X/2), int32(pnt.coord.Size.Y/2), pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	}
}

type PaintPoly struct {
	x     []int16
	y     []int16
	cd    OsCd
	thick int
}

func (pnt *PaintPoly) Add(pos OsV2) {
	pnt.x = append(pnt.x, int16(pos.X))
	pnt.y = append(pnt.y, int16(pos.Y))
}

func (pnt *PaintPoly) Draw(ui *Ui) {

	if pnt.thick == 0 {
		gfx.FilledPolygonRGBA(ui.render, ui.poly.x, ui.poly.y, pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	} else {
		gfx.AAPolygonRGBA(ui.render, ui.poly.x, ui.poly.y, pnt.cd.R, pnt.cd.G, pnt.cd.B, pnt.cd.A)
	}
}

type PaintImage struct {
	path        ResourcePath
	inverserRGB bool
	coord       OsV4
	cd          OsCd
}

func PaintImage_load(path ResourcePath, inverserRGB bool, ui *Ui) *Image {

	var img *Image
	for _, it := range ui.images {
		if it.path.Cmp(&path) && it.inverserRGB == inverserRGB {
			img = it
			break
		}
	}

	if img == nil {
		var err error
		img, err = NewImage(path, inverserRGB, ui.render)
		if err != nil {
			return nil //fmt.Errorf("NewImage() failed: %w", err)
		}

		ui.images = append(ui.images, img)
	}

	return img
}
func (pnt *PaintImage) Draw(ui *Ui) {

	img := PaintImage_load(pnt.path, pnt.inverserRGB, ui)

	if img == nil {
		return //errors.New("image not found")
	}

	err := img.Draw(pnt.coord, pnt.cd, ui.render)
	if err != nil {
		return //fmt.Errorf("Draw() failed: %w", err)
	}
}

type PaintText struct {
	text  string
	coord OsV4
	font  *Font
	cd    OsCd
	h     int
	align OsV2
	cds   []OsCd
}

func (pnt *PaintText) Draw(ui *Ui) {

	err := pnt.font.Print(pnt.text, pnt.h, pnt.coord, pnt.align, pnt.cd, pnt.cds, ui.render)
	if err != nil {
		return //fmt.Errorf("Text().Print() failed: %w", err)
	}
}

type PaintBuff struct {
	ui    *Ui
	items []interface{}

	lastCrop OsV4
}

func NewPaintBuff(ui *Ui) *PaintBuff {
	var b PaintBuff
	b.ui = ui
	return &b
}

func (b *PaintBuff) Reset(crop OsV4) {
	b.lastCrop = crop
	b.items = b.items[:0]
}

func (b *PaintBuff) Add(it interface{}) {
	b.items = append(b.items, it)
}

func (b *PaintBuff) AddCrop(coord OsV4) OsV4 {

	b.Add(PaintCrop{coord: coord})

	old := b.lastCrop
	b.lastCrop = coord
	return old
}

func (b *PaintBuff) AddRect(coord OsV4, cd OsCd, thick int) {
	b.Add(PaintRect{coord: coord, cd: cd, thick: thick})
}

func (b *PaintBuff) AddLine(start OsV2, end OsV2, cd OsCd, thick int) {
	b.Add(PaintLine{start: start, end: end, cd: cd, thick: thick})
}

func (b *PaintBuff) AddCircle(coord OsV4, cd OsCd, thick int) {
	b.Add(PaintCircle{coord: coord, cd: cd, thick: thick})
}

func (b *PaintBuff) AddImage(path ResourcePath, inverserRGB bool, coord OsV4, cd OsCd, alignV int, alignH int, fill bool) {

	img := PaintImage_load(path, inverserRGB, b.ui)

	var q OsV4
	if !fill {
		rect_size := OsV2_InRatio(coord.Size, img.origSize)
		q = CQuad_center(coord, rect_size)
	} else {
		q.Start = coord.Start
		q.Size = OsV2_OutRatio(coord.Size, img.origSize)
	}

	if alignH == 0 {
		q.Start.X = coord.Start.X
	} else if alignH == 1 {
		q.Start.X = CQuad_centerFull(coord, q.Size).Start.X
	} else if alignH == 2 {
		q.Start.X = coord.End().X - q.Size.X
	}

	if alignV == 0 {
		q.Start.Y = coord.Start.Y
	} else if alignV == 1 {
		q.Start.Y = CQuad_centerFull(coord, q.Size).Start.Y
	} else if alignV == 2 {
		q.Start.Y = coord.End().Y - q.Size.Y
	}

	imgRectBackup := b.AddCrop(b.lastCrop.GetIntersect(coord))

	b.Add(PaintImage{path: path, inverserRGB: inverserRGB, cd: cd, coord: q})

	b.AddCrop(imgRectBackup)
}

func (b *PaintBuff) PaintText(text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, cds []OsCd) {

	b.Add(PaintText{text: text, coord: coord, font: font, cd: cd, h: h, align: align, cds: cds})
}

func (b *PaintBuff) PaintTextBack(rangee OsV2, text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, underline bool, addSpaceY bool) error {

	if rangee.X == rangee.Y {
		return nil
	}

	start, err := font.Start(text, h, coord, align, nil)
	if err != nil {
		return fmt.Errorf("Start() failed: %w", err)
	}

	var rng OsV2
	rng.X, err = font.GetPxPos(text, h, rangee.X)
	if err != nil {
		return fmt.Errorf("GetPxPos(1) failed: %w", err)
	}
	rng.Y, err = font.GetPxPos(text, h, rangee.Y)
	if err != nil {
		return fmt.Errorf("GetPxPos(2) failed: %w", err)
	}
	rng.Sort()

	if rng.X != rng.Y {
		if underline {
			Y := coord.Start.Y + coord.Size.Y
			b.AddRect(OsV4{Start: OsV2{start.X + rng.X, Y - 2}, Size: OsV2{rng.Y, 2}}, cd, 0)
			//_Ui_boxSE(ui.render, OsV2{start.X + rng.X, Y - 2}, OsV2{start.X + rng.Y, Y}, cd)
		} else {
			c := InitOsQuad(start.X+rng.X, coord.Start.Y, rng.Y-rng.X, coord.Size.Y)
			if addSpaceY {
				c = c.AddSpaceY((coord.Size.Y - h) / 4) //smaller height
			}
			b.AddRect(c, cd, 0)
			//_Ui_boxSE(ui.render, c.Start, c.End(), cd)

		}
	}
	return nil
}

func (b *PaintBuff) PaintTextCursor(text string, coord OsV4, font *Font, cd OsCd, h int, align OsV2, cursorPos int, cell int) (OsV4, error) {

	b.ui.cursorEdit = true
	cd.A = b.ui.cursorCdA

	start, err := font.Start(text, h, coord, align, nil)
	if err != nil {
		return OsV4{}, fmt.Errorf("TextCursor().Start() failed: %w", err)
	}

	ex, err := font.GetPxPos(text, h, cursorPos)
	if err != nil {
		return OsV4{}, fmt.Errorf("TextCursor().GetPxPos() failed: %w", err)
	}

	cursorQuad := InitOsQuad(start.X+ex, coord.Start.Y, OsMax(1, cell/15), coord.Size.Y)
	cursorQuad = cursorQuad.AddSpaceY((coord.Size.Y - h) / 4) //smaller height

	b.AddRect(cursorQuad, cd, 0)

	return cursorQuad, nil
}

func (b *PaintBuff) Draw() {

	ui := b.ui

	for _, it := range b.items {
		switch v := it.(type) {
		case PaintCrop:
			v.Draw(ui)
		case PaintRect:
			v.Draw(ui)
		case PaintLine:
			v.Draw(ui)
		case PaintCircle:
			v.Draw(ui)
		case PaintPoly:
			v.Draw(ui)
		case PaintImage:
			v.Draw(ui)
		case PaintText:
			v.Draw(ui)
		}
	}
}
*/
