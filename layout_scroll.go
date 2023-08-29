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

type LayerScroll struct {
	wheel int // pixel move

	data_height   int
	screen_height int

	clickRel int

	timeWheel int

	show   bool
	narrow bool
}

func (scroll *LayerScroll) Init() {

	scroll.clickRel = 0
	scroll.wheel = 0
	scroll.data_height = 1
	scroll.screen_height = 1
	scroll.timeWheel = 0
	scroll.show = true
	scroll.narrow = false
}

func (scroll *LayerScroll) GetWheel() int {

	if scroll.data_height > scroll.screen_height {
		return OsClamp(scroll.wheel, 0, (scroll.data_height - scroll.screen_height))
	}
	return 0
}

func (scroll *LayerScroll) SetWheel(wheelPixel int) bool {

	oldWheel := scroll.wheel

	scroll.wheel = wheelPixel
	scroll.wheel = scroll.GetWheel() //clamp by boundaries

	/*if scroll.data_height > scroll.screen_height {
		scroll.wheel = OsClamp(wheelPixel, 0, (scroll.data_height - scroll.screen_height))
	} else {
		scroll.wheel = 0
	}*/

	/*if scroll.data_height == 0 { // otherwise saved scroll is reset!
		scroll.wheel = oldWheel
	}*/

	if oldWheel != scroll.wheel {
		scroll.timeWheel = OsTicks()
	}

	return oldWheel != scroll.wheel
}

func (scroll *LayerScroll) Is() bool {

	return scroll.show && scroll.data_height > scroll.screen_height
}

// shorter bool,
func (scroll *LayerScroll) GetScrollBackCoordV(coord OsV4, ui *Ui) OsV4 {
	WIDTH := scroll._GetWidth(ui)
	H := 0 // OsTrn(shorter, WIDTH, 0)
	return OsV4{OsV2{coord.Start.X + coord.Size.X, coord.Start.Y}, OsV2{WIDTH, scroll.screen_height - H}}
}
func (scroll *LayerScroll) GetScrollBackCoordH(coord OsV4, ui *Ui) OsV4 {
	WIDTH := scroll._GetWidth(ui)
	H := 0 //OsTrn(shorter, WIDTH, 0)
	return OsV4{OsV2{coord.Start.X, coord.Start.Y + coord.Size.Y}, OsV2{scroll.screen_height - H, WIDTH}}
}

func (scroll *LayerScroll) _GetWidth(ui *Ui) int {
	widthWin := ui.Cell() / 2
	if scroll.narrow {
		return OsMax(4, widthWin/10)
	}
	return widthWin
}

func (scroll *LayerScroll) _UpdateV(coord OsV4, ui *Ui) OsV4 {

	var outSlider OsV4
	if scroll.data_height <= scroll.screen_height {
		outSlider.Start = coord.Start

		outSlider.Size.X = scroll._GetWidth(ui)
		outSlider.Size.Y = coord.Size.Y // self.screen_height
	} else {
		outSlider.Start.X = coord.Start.X
		outSlider.Start.Y = coord.Start.Y + int(float32(coord.Size.Y)*(float32(scroll.GetWheel())/float32(scroll.data_height)))

		outSlider.Size.X = scroll._GetWidth(ui)
		outSlider.Size.Y = int(float32(coord.Size.Y) * (float32(scroll.screen_height) / float32(scroll.data_height)))
	}
	return outSlider
}

func (scroll *LayerScroll) _UpdateH(start OsV2, ui *Ui) OsV4 {

	var outSlider OsV4
	if scroll.data_height <= scroll.screen_height {
		outSlider.Start = start

		outSlider.Size.X = scroll.screen_height
		outSlider.Size.Y = scroll._GetWidth(ui)
	} else {
		outSlider.Start.X = start.X + int(float32(scroll.screen_height)*(float32(scroll.GetWheel())/float32(scroll.data_height)))
		outSlider.Start.Y = start.Y

		outSlider.Size.X = int(float32(scroll.screen_height) * (float32(scroll.screen_height) / float32(scroll.data_height)))
		outSlider.Size.Y = scroll._GetWidth(ui)
	}
	return outSlider
}

func (scroll *LayerScroll) _GetSlideCd(ui *Ui) OsCd {

	cd_slide := ui.io.GetThemeCd()
	if scroll.data_height <= scroll.screen_height {
		cd_slide = OsCd_Aprox(OsCd_themeBack(), cd_slide, 0.5) // disable
	}

	return cd_slide
}

func (scroll *LayerScroll) DrawV(coord OsV4, showBackground bool, buff *PaintBuff) {
	slider := scroll._UpdateV(coord, buff.ui)

	slider = slider.AddSpace(OsMax(1, slider.Size.X/5))

	// make scroll visible if there is a lot of records(items)
	if slider.Size.Y == 0 {
		c := buff.ui.Cell() / 4
		slider.Start.Y -= c / 2
		slider.Size.Y += c
	}

	if showBackground {
		cdB := OsCd_black()
		cdB.A = 30
		buff.AddRect(coord, cdB, 0)
	}
	buff.AddRect(slider, scroll._GetSlideCd(buff.ui), 0)
}

func (scroll *LayerScroll) DrawH(coord OsV4, showBackground bool, buff *PaintBuff) {

	slider := scroll._UpdateH(coord.Start, buff.ui)

	slider = slider.AddSpace(OsMax(1, slider.Size.Y/5))

	// make scroll visible if there is a lot of records(items)
	if slider.Size.Y == 0 {

		c := buff.ui.Cell() / 4
		slider.Start.X -= c / 2
		slider.Size.X += c
	}

	if showBackground {
		cdB := OsCd_black()
		cdB.A = 30
		buff.AddRect(coord, cdB, 0)
	}
	buff.AddRect(slider, scroll._GetSlideCd(buff.ui), 0)
}

func (scroll *LayerScroll) _GetTempScroll(srcl int, ui *Ui) int {

	return ui.Cell() * srcl
}

func (scroll *LayerScroll) TouchV(parentCoord OsV4, scrollCoord OsV4, isTouched bool, root *Root) bool {

	ui := root.ui

	//scroll.SetWheel(scroll.wheel) //refresh

	if scrollCoord.Inside(ui.io.touch.pos) {
		ui.PaintCursor("default")
	}

	insideParent := parentCoord.Inside(ui.io.touch.pos)
	//fmt.Println(insideParent)
	if ui.io.touch.wheel && insideParent && !ui.io.keys.shift {
		if scroll.SetWheel(scroll.GetWheel() + scroll._GetTempScroll(ui.io.touch.wheelPos, ui)) {
			ui.io.touch.wheelPos = 0 // let parent scroll
		}
	}

	if !root.touch.IsAnyActive() && !ui.io.keys.shift && insideParent {
		if ui.io.keys.arrowU {
			if scroll.SetWheel(scroll.GetWheel() - ui.Cell()) {
				ui.io.keys.arrowU = false
			}
		}
		if ui.io.keys.arrowD {
			if scroll.SetWheel(scroll.GetWheel() + ui.Cell()) {
				ui.io.keys.arrowD = false
			}
		}

		if ui.io.keys.home {
			if scroll.SetWheel(0) {
				ui.io.keys.home = false
			}
		}
		if ui.io.keys.end {
			if scroll.SetWheel(scroll.data_height) {
				ui.io.keys.end = false
			}
		}

		if ui.io.keys.pageU {
			if scroll.SetWheel(scroll.GetWheel() - scroll.screen_height) {
				ui.io.keys.pageU = false
			}
		}
		if ui.io.keys.pageD {
			if scroll.SetWheel(scroll.GetWheel() + scroll.screen_height) {
				ui.io.keys.pageD = false
			}
		}
	}

	if !scroll.Is() {
		return false
	}

	sliderFront := scroll._UpdateV(scrollCoord, ui)
	midSlider := sliderFront.Size.Y / 2

	inside := sliderFront.Inside(ui.io.touch.pos)
	if ui.io.touch.start {
		isTouched = inside
		scroll.clickRel = ui.io.touch.pos.Y - sliderFront.Start.Y - midSlider // rel to middle of front slide
	}

	if isTouched { // click on slider
		mid := float32((ui.io.touch.pos.Y - scrollCoord.Start.Y) - midSlider - scroll.clickRel)
		scroll.SetWheel(int((mid / float32(scrollCoord.Size.Y)) * float32(scroll.data_height)))

	} else if ui.io.touch.start && scrollCoord.Inside(ui.io.touch.pos) && !sliderFront.Inside(ui.io.touch.pos) { // click(once) on background
		mid := float32((ui.io.touch.pos.Y - scrollCoord.Start.Y) - midSlider)
		scroll.SetWheel(int((mid / float32(scrollCoord.Size.Y)) * float32(scroll.data_height)))

		// switch to 'click on slider'
		isTouched = true
		scroll.clickRel = 0
	}

	return isTouched
}

func (scroll *LayerScroll) TouchH(parentCoord OsV4, scrollCoord OsV4, needShiftWheel bool, isTouched bool, root *Root) bool {
	ui := root.ui

	//scroll.SetWheel(scroll.wheel) //refresh

	if scrollCoord.Inside(ui.io.touch.pos) {
		ui.PaintCursor("default")
	}

	insideParent := parentCoord.Inside(ui.io.touch.pos)
	if ui.io.touch.wheel && insideParent && (!needShiftWheel || ui.io.keys.shift) {
		if scroll.SetWheel(scroll.GetWheel() + scroll._GetTempScroll(ui.io.touch.wheelPos, ui)) {
			ui.io.touch.wheelPos = 0 // let parent scroll
		}
	}

	if !root.touch.IsAnyActive() && (!needShiftWheel || ui.io.keys.shift) && insideParent {
		if ui.io.keys.arrowL {
			if scroll.SetWheel(scroll.GetWheel() - ui.Cell()) {
				ui.io.keys.arrowL = false
			}
		}
		if ui.io.keys.arrowR {
			if scroll.SetWheel(scroll.GetWheel() + ui.Cell()) {
				ui.io.keys.arrowR = false
			}
		}

		if ui.io.keys.home {
			if scroll.SetWheel(0) {
				ui.io.keys.home = false
			}
		}
		if ui.io.keys.end {
			if scroll.SetWheel(scroll.data_height) {
				ui.io.keys.end = false
			}
		}

		if ui.io.keys.pageU {
			if scroll.SetWheel(scroll.GetWheel() - scroll.screen_height) {
				ui.io.keys.pageU = false
			}
		}
		if ui.io.keys.pageD {
			if scroll.SetWheel(scroll.GetWheel() + scroll.screen_height) {
				ui.io.keys.pageD = false
			}
		}
	}

	if !scroll.Is() {
		return false
	}

	sliderFront := scroll._UpdateH(scrollCoord.Start, ui)
	midSlider := sliderFront.Size.X / 2

	inside := sliderFront.Inside(ui.io.touch.pos)
	if ui.io.touch.start {
		isTouched = inside
		scroll.clickRel = ui.io.touch.pos.X - sliderFront.Start.X - midSlider // rel to middle of front slide
	}

	if isTouched { // click on slider
		mid := float32((ui.io.touch.pos.X - scrollCoord.Start.X) - midSlider - scroll.clickRel)
		scroll.SetWheel(int((mid / float32(scroll.screen_height)) * float32(scroll.data_height)))
	} else if ui.io.touch.start && scrollCoord.Inside(ui.io.touch.pos) && !sliderFront.Inside(ui.io.touch.pos) { // click(once) on background
		mid := float32((ui.io.touch.pos.X - scrollCoord.Start.X) - midSlider)
		scroll.SetWheel(int((mid / float32(scroll.screen_height)) * float32(scroll.data_height)))

		// switch to 'click on slider'
		isTouched = true
		scroll.clickRel = 0
	}

	return isTouched
}

func (scroll *LayerScroll) TryDragScroll(fast_dt int, sign int, ui *Ui) bool {
	wheelOld := scroll.GetWheel()

	dt := int((1.0 / 2.0) / float32(fast_dt) * 1000)

	if OsTicks()-scroll.timeWheel > dt {
		scroll.SetWheel(scroll.GetWheel() + scroll._GetTempScroll(sign, ui))
	}

	return scroll.GetWheel() != wheelOld
}
