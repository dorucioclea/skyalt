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

type Layout struct {
	cols LayoutArray
	rows LayoutArray

	scrollV        LayerScroll
	scrollH        LayerScroll
	scrollOnScreen bool //show scroll all the time

	//maybe have them one time in Root like a 'over* Layout', 'overScroll* Layout', etc. ...
	over          bool
	overScroll    bool
	touch_inside  bool
	touch_active  bool
	touch_end     bool
	touch_enabled bool
}

func (lay *Layout) Init(hash uint64, infoLayout *RS_LScroll) {

	lay.touch_enabled = true

	lay.scrollV.Init()
	lay.scrollH.Init()

	it := _Layout_findGlobalScrollHash(infoLayout, hash)
	if it != nil {

		lay.scrollV.wheel = it.ScrollVpos
		lay.scrollH.wheel = it.ScrollHpos

		for _, rs := range it.Cols_resize {
			//pos := OsMax(0, rs.Pos)
			res, _ := lay.cols.FindOrAddResize(rs.Name)
			res.value = float32(rs.Value)
			//lay.cols.Resize(pos + 1)
			//lay.cols.items[pos].resize = float32(rs.Value)
		}

		for _, rs := range it.Rows_resize {
			//pos := OsMax(0, rs.Pos)
			res, _ := lay.rows.FindOrAddResize(rs.Name)
			res.value = float32(rs.Value)
			//lay.rows.Resize(pos + 1)
			//lay.rows.items[pos].resize = float32(rs.Value)
		}
	}
}

func (lay *Layout) Save(hash uint64, infoLayout *RS_LScroll) {

	hasColResize := lay.cols.HasResize()
	hasRowResize := lay.rows.HasResize()

	// save scroll into Rec
	if lay.scrollV.wheel != 0 || lay.scrollH.wheel != 0 || hasColResize || hasRowResize {
		it := _Layout_addGlobalScrollHash(infoLayout, hash)

		it.ScrollVpos = 0
		it.ScrollHpos = 0
		it.Cols_resize = nil
		it.Rows_resize = nil

		if lay.scrollV.wheel != 0 {
			it.ScrollVpos = lay.scrollV.wheel
		}
		if lay.scrollH.wheel != 0 {
			it.ScrollHpos = lay.scrollH.wheel
		}

		if hasColResize {
			for _, c := range lay.cols.resizes {
				it.Cols_resize = append(it.Cols_resize, RS_LResize{Name: c.name, Value: c.value})
			}
		}

		if hasRowResize {
			for _, r := range lay.rows.resizes {
				it.Rows_resize = append(it.Rows_resize, RS_LResize{Name: r.name, Value: r.value})
			}
		}

	} else {
		sc := _Layout_findGlobalScrollHash(infoLayout, hash)
		if sc != nil {
			*sc = RS_LScrollItem{}
		}
	}
}

func (lay *Layout) Close(hash uint64, infoLayout *RS_LScroll) {
	lay.Save(hash, infoLayout)
}

func (lay *Layout) Reset() {
	lay.cols.Clear()
	lay.rows.Clear()
}

func (lay *Layout) UpdateArray(cell int, window OsV2, endGrid OsV2) {

	if endGrid.X > lay.cols.NumIns() {
		lay.cols.Resize(int(endGrid.X))
	}
	if endGrid.Y > lay.rows.NumIns() {
		lay.rows.Resize(int(endGrid.Y))
	}
	lay.cols.Update(cell, window.X)
	lay.rows.Update(cell, window.Y)
}

func (lay *Layout) Convert(cell int, in OsV4) OsV4 {

	c := lay.cols.Convert(cell, in.Start.X, in.Start.X+in.Size.X)
	r := lay.rows.Convert(cell, in.Start.Y, in.Start.Y+in.Size.Y)

	return OsV4{OsV2{c.X, r.X}, OsV2{c.Y, r.Y}}
}

func (lay *Layout) ConvertMax(cell int, in OsV4) OsV4 {
	c := lay.cols.ConvertMax(cell, in.Start.X, in.Start.X+in.Size.X)
	r := lay.rows.ConvertMax(cell, in.Start.Y, in.Start.Y+in.Size.Y)

	return OsV4{OsV2{c.X, r.X}, OsV2{c.Y, r.Y}}
}

func _Layout_findGlobalScrollHash(scroll *RS_LScroll, hash uint64) *RS_LScrollItem {

	if scroll == nil {
		return nil
	}

	for _, it := range scroll.items {
		if it.Hash == hash {
			return it
		}
	}

	return nil
}

func _Layout_addGlobalScrollHash(scroll *RS_LScroll, hash uint64) *RS_LScrollItem {
	if scroll == nil {
		return nil
	}

	sc := _Layout_findGlobalScrollHash(scroll, hash)
	if sc != nil {
		return sc
	}

	nw := &RS_LScrollItem{Hash: hash}
	scroll.items = append(scroll.items, nw)
	return nw
}
