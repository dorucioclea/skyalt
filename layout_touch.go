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

type LayerTouch struct {
	canvas  *LayoutDiv
	scrollV *LayoutDiv
	scrollH *LayoutDiv
	resize  *LayoutDiv
}

func (layTouch *LayerTouch) Set(canvas *LayoutDiv, scrollV *LayoutDiv, scrollH *LayoutDiv, resize *LayoutDiv) {
	layTouch.canvas = canvas
	layTouch.scrollV = scrollV
	layTouch.scrollH = scrollH
	layTouch.resize = resize
}

func (layTouch *LayerTouch) Reset() {
	*layTouch = LayerTouch{}
}

func (layTouch *LayerTouch) IsAnyActive() bool {
	return layTouch.canvas != nil || layTouch.scrollV != nil || layTouch.scrollH != nil || layTouch.resize != nil
}

func (layTouch *LayerTouch) IsScrollOrResizeActive() bool {
	return layTouch.scrollV != nil || layTouch.scrollH != nil || layTouch.resize != nil
}

func (layTouch *LayerTouch) IsFnMove(canvas *LayoutDiv, scrollV *LayoutDiv, scrollH *LayoutDiv, resize *LayoutDiv) bool {
	return ((canvas != nil && layTouch.canvas == canvas) || (scrollV != nil && layTouch.scrollV == scrollV) || (scrollH != nil && layTouch.scrollH == scrollH) || (resize != nil && layTouch.resize == resize))
}
