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

type LayoutLevel struct {
	parent *LayoutLevel

	name string
	use  bool

	buff *PaintBuff

	infoLayout *RS_LScroll

	div      *LayoutDiv //rootPack
	openPack *LayoutDiv
	stack    *LayoutDiv

	src_coordMoveCut OsV4

	next *LayoutLevel

	close bool
}

func NewLayoutLevel(parent *LayoutLevel, name string, src_coordMoveCut OsV4, openPack *LayoutDiv, infoLayout *RS_LScroll, ui *Ui) *LayoutLevel {

	var self LayoutLevel

	self.parent = parent
	self.use = true
	self.name = name

	self.buff = NewPaintBuff(ui)

	self.infoLayout = infoLayout

	self.div = NewLayoutPack(nil, "", OsV4{}, infoLayout)

	self.openPack = openPack

	self.src_coordMoveCut = src_coordMoveCut

	return &self
}

func (level *LayoutLevel) Free() {
	level.div.Destroy(level.infoLayout)
}

func (level *LayoutLevel) Delete() {
	if level.next != nil {
		level.next.Delete()
	}

	level.Free()

	if level.parent != nil {
		level.parent.next = nil
	}
}

func (level *LayoutLevel) GetCoord(q OsV4, winRect OsV4) OsV4 {

	if !level.src_coordMoveCut.IsZero() {
		// relative
		q = OsV4_relativeSurround(level.src_coordMoveCut, q, winRect)
	} else {
		// center
		q = OsV4_center(winRect, q.Size)
	}
	return q
}

func (level *LayoutLevel) IsTop() bool {
	return level.next == nil
}

func (level *LayoutLevel) AddLevel(name string, src_coordMoveCut OsV4, openPack *LayoutDiv, ui *Ui) *LayoutLevel {

	if level.next != nil {
		if level.next.name == name {
			return level.next
		}
	}

	level.next = NewLayoutLevel(level, name, src_coordMoveCut, openPack, level.infoLayout, ui)
	return level.next
}
