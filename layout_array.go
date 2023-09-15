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

type LayoutArrayItem struct {
	min    float32
	max    float32
	resize *LayoutArrayRes
}

type LayoutArrayRes struct {
	value float32
	name  string
}

type LayoutArray struct {
	resizes []*LayoutArrayRes //backups

	items   []LayoutArrayItem
	outputs []int32
}

func (arr *LayoutArray) Clear() {
	arr.items = arr.items[:0]
}

func (arr *LayoutArray) NumIns() int {
	return len(arr.items)
}

func (arr *LayoutArray) Resize(num int) {
	//huge amount of RAM if I set few items on position 1000 - all before are alocated and set & reset after frame is rendered ...
	for i := arr.NumIns(); i < num; i++ {
		arr.items = append(arr.items, LayoutArrayItem{min: 1, max: 0, resize: nil})
	}
}

func LayoutArray_resizerSize(cell int) int {
	v := cell / 4
	if v < 9 {
		return 9
	}
	return v
}

func (arr *LayoutArray) IsLastResizeValid() bool {
	n := arr.NumIns()
	return n >= 2 && arr.items[n-2].resize == nil && arr.items[n-1].resize != nil
}

func (arr *LayoutArray) GetResizeIndex(i int) int {

	if arr.IsLastResizeValid() {
		if i+2 == arr.NumIns() {
			if arr.items[i+1].resize != nil {
				return i + 1 // show resizer before column/row
			}
			return -1
		}
		if i+1 == arr.NumIns() {
			return -1 // last was return as Previous, so no last
		}
	}

	if i < arr.NumIns() {
		if arr.items[i].resize != nil {
			return i
		}
	}
	return -1
}

func (arr *LayoutArray) Convert(cell int, start int, end int) OsV2 {

	var ret OsV2

	for i := 0; i < end; i++ {
		ok := (i < len(arr.outputs))

		if i < start {
			if ok {
				ret.X += int(arr.outputs[i])
			} else {
				ret.X += cell
			}
		} else {
			if ok {
				ret.Y += int(arr.outputs[i])
			} else {
				ret.Y += cell
			}
		}
	}

	if end > 0 && (end-1 < arr.NumIns()) && arr.GetResizeIndex(int(end)-1) >= 0 {
		ret.Y -= LayoutArray_resizerSize(cell)
	}

	return ret
}

func (arr *LayoutArray) ConvertMax(cell int, start int, end int) OsV2 {
	var ret OsV2

	for i := 0; i < end; i++ {
		ok := (i < arr.NumIns())

		if i < start {
			if ok {
				ret.X += int(arr.items[i].max * float32(cell))
			} else {
				ret.X += cell
			}
		} else {
			if ok {
				ret.Y += int(arr.items[i].max * float32(cell))
			} else {
				ret.Y += cell
			}
		}
	}

	return ret
}

func (arr *LayoutArray) GetCloseCell(pos int) int {
	if pos < 0 {
		return -1
	}
	allPixels := 0
	allPixelsLast := 0
	for i := 0; i < len(arr.outputs); i++ {
		allPixels += int(arr.outputs[i])

		if pos >= allPixelsLast && pos < allPixels {
			return i
		}

		allPixelsLast = allPixels
	}

	return len(arr.outputs)
}

func (arr *LayoutArray) GetResizerPos(i int, cell int) int {
	if i >= len(arr.outputs) {
		return 0
	}

	allPixels := 0
	for ii := 0; ii <= i; ii++ {
		allPixels += int(arr.outputs[ii])
	}

	return allPixels - LayoutArray_resizerSize(cell)
}

func (arr *LayoutArray) IsResizerTouch(touchPos int, cell int) int {
	space := LayoutArray_resizerSize(cell)

	for i := 0; i < arr.NumIns(); i++ {
		if arr.GetResizeIndex(i) >= 0 {
			p := arr.GetResizerPos(i, cell)
			if touchPos > p && touchPos < p+space {
				return i
			}
		}
	}
	return -1
}

func (arr *LayoutArray) OutputAll() int {
	sum := 0
	for i := 0; i < len(arr.outputs); i++ {
		sum += int(arr.outputs[i])
	}
	return sum
}

func (arr *LayoutArray) Update(cell int, window int) {

	arr.outputs = make([]int32, arr.NumIns())

	//project in -> out
	//for _, it := range arr.items {
	for i := 0; i < len(arr.items); i++ {

		//min
		minV := float64(arr.items[i].min)
		minV = OsClampFloat(minV, 0.001, 100000000)

		if arr.items[i].resize == nil {
			// max
			maxV := minV
			if arr.items[i].max > 0 {
				maxV = float64(arr.items[i].max)
				maxV = OsMaxFloat(minV, maxV)
			}

			arr.items[i].min = float32(minV)
			arr.items[i].max = float32(maxV)
		} else {

			resV := float64(arr.items[i].resize.value)
			resV = OsMaxFloat(resV, minV)

			if arr.items[i].max > 0 {
				maxV := float64(arr.items[i].max)
				maxV = OsMaxFloat(minV, maxV)

				resV = OsClampFloat(resV, minV, maxV)
			}

			arr.items[i].min = float32(resV)
			arr.items[i].max = float32(resV)
			arr.items[i].resize.value = float32(resV)
		}
	}

	//sum
	allPixels := 0
	for i := 0; i < len(arr.outputs); i++ {
		arr.outputs[i] = int32(arr.items[i].min * float32(cell))
		allPixels += int(arr.outputs[i])
	}

	// make it larger(if maxes allow)
	hasSpace := (len(arr.outputs) > 0)
	for allPixels < window && hasSpace {
		rest := window - allPixels
		tryAdd := OsMax(1, rest/int(len(arr.outputs)))

		hasSpace = false
		for i := 0; i < len(arr.outputs) && allPixels < window; i++ {

			maxAdd := int(arr.items[i].max*float32(cell)) - int(arr.outputs[i])
			add := OsClamp(tryAdd, 0, maxAdd)

			arr.outputs[i] += int32(add)
			allPixels += add

			if maxAdd > tryAdd {
				hasSpace = true
			}
		}
	}
}

func (arr *LayoutArray) HasResize() bool {
	for _, c := range arr.items {
		if c.resize != nil {
			return true
		}
	}
	return false
}

func (arr *LayoutArray) FindOrAddResize(name string) (*LayoutArrayRes, bool) {

	//find
	for _, it := range arr.resizes {
		if it.name == name {
			return it, true
		}
	}

	//add
	it := &LayoutArrayRes{name: name, value: 1}
	arr.resizes = append(arr.resizes, it)
	return it, false
}

func (arr *LayoutArray) GetOutput(i int) int {

	if i < len(arr.outputs) {
		return int(arr.outputs[i])
	}
	return -1
}

func (arr *LayoutArray) findOrAdd(pos int) *LayoutArrayItem {

	if pos >= len(arr.items) {
		arr.Resize(pos + 1)
	}

	return &arr.items[pos]
}
