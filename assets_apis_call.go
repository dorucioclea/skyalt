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
	"errors"
	"fmt"
)

const TpI64 = byte(0x7e)
const TpF32 = byte(0x7d)
const TpF64 = byte(0x7c)
const TpBytes = byte(0x7b)
const TpString = byte(0x7a)

func (asset *Asset) fn_call(assetName string, fnName string, args []byte) (int64, error) {

	if len(fnName) == 0 {
		return -1, fmt.Errorf("'fnName' is empty")
	}

	ass := asset.findAsset(assetName)
	if ass == nil {
		return -1, fmt.Errorf("Asset(%s) not found", assetName)
	}

	return ass.Call(fnName, args)
}

func (asset *Asset) _sa_fn_call(assetMem uint64, fnMem uint64, argsMem uint64) int64 {
	args, err := asset.ptrToBytesDirect(argsMem)
	if asset.AddLogErr(err) {
		return -1
	}

	assetName, err := asset.ptrToString(assetMem)
	if asset.AddLogErr(err) {
		return -1
	}
	fnName, err := asset.ptrToString(fnMem)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.fn_call(assetName, fnName, args)
	if asset.AddLogErr(err) {
		return -1
	}

	return ret
}

func (asset *Asset) fn_setReturn(args []byte) int64 {
	//clone
	asset.app.fn2Returns = make([]byte, len(args))
	copy(asset.app.fn2Returns, args)
	return 1
}
func (asset *Asset) _sa_fn_setReturn(argsMem uint64) int64 {
	args, err := asset.ptrToBytesDirect(argsMem)
	if asset.AddLogErr(err) {
		return -1
	}
	return asset.fn_setReturn(args)
}

func (asset *Asset) fn_getReturn() []byte {
	v := asset.app.fn2Return
	v = append(v, asset.app.fn2Returns...)
	return v
}
func (asset *Asset) _sa_fn_getReturn(argsMem uint64) int64 {
	if asset.wasm == nil {
		asset.AddLogErr(errors.New("no wasm module or in debug mode"))
		return -1
	}

	ptr, _ := _ptrg(argsMem)

	if !asset.wasm.mod.Memory().Write(ptr, asset.app.fn2Return) {
		asset.AddLogErr(errors.New("write1 to wasm failed"))
		return -1
	}
	ptr += uint32(len(asset.app.fn2Return))

	//returns(multiple)
	if !asset.wasm.mod.Memory().Write(ptr, asset.app.fn2Returns) {
		asset.AddLogErr(errors.New("write2 to wasm failed"))
		return -1
	}

	return 1
}
