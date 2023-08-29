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

type LayerTile struct {
	coord OsV4
	text  string
	cd    OsCd

	ticks          int
	needRedraw     bool
	tickOpenForSet bool
}

func (layTile *LayerTile) NextTick() {
	layTile.tickOpenForSet = true
}

func (layTile *LayerTile) timeToShow() int {
	return layTile.ticks + 200 - OsTicks()
}

func (layTile *LayerTile) IsActive(touchPos OsV2) bool {

	if !layTile.coord.Inside(touchPos) && len(layTile.text) > 0 {
		*layTile = LayerTile{} //reset
	}

	draw := (len(layTile.text) > 0 && layTile.timeToShow() <= 0)
	if draw {
		layTile.needRedraw = false
	}
	return draw
}

func (layTile *LayerTile) NeedsRedrawFromSleep(touchPos OsV2) bool {

	redraw := layTile.needRedraw
	if !layTile.IsActive(touchPos) {
		redraw = false
	}
	return redraw
}

func (layTile *LayerTile) Set(touchPos OsV2, coord OsV4, text string, cd OsCd) {

	if coord.Inside(touchPos) {
		if layTile.tickOpenForSet && (!layTile.coord.Cmp(coord) || layTile.text != text) {
			layTile.coord = coord
			layTile.cd = cd

			layTile.text = text

			layTile.ticks = OsTicks()
			layTile.needRedraw = true
		}

		layTile.tickOpenForSet = false
	}
}
