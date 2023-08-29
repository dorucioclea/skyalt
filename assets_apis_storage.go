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
	"strings"
)

func _ptrg(mem uint64) (uint32, uint32) {
	//endiness ...
	ptr := uint32(mem >> 32)
	size := uint32(mem)
	return ptr, size
}

func (asset *Asset) ptrToString(mem uint64) (string, error) {
	if asset.wasm == nil {
		return "", errors.New("wasm is nil")
	}

	ptr, size := _ptrg(mem)

	bytes, ok := asset.wasm.mod.Memory().Read(ptr, size)
	if !ok {
		return "", fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, asset.wasm.mod.Memory().Size())
	}
	return strings.Clone(string(bytes)), nil
}

func (asset *Asset) stringToPtr(str string, dst uint64) error {
	if asset.wasm == nil {
		return errors.New("wasm is nil")
	}

	ptr, size := _ptrg(dst)
	n := uint32(len(str))
	if n < size {
		size = n
	}

	if !asset.wasm.mod.Memory().Write(ptr, []byte(str)) {
		return fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, asset.wasm.mod.Memory().Size())
	}
	return nil
}

/*func (asset *Asset) ptrToBytes(mem uint64) ([]byte, error) {
	if asset.wasm == nil {
		return nil, errors.New("wasm is nil")
	}

	ptr, size := _ptrg(mem)
	bts, ok := asset.wasm.mod.Memory().Read(ptr, size)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, asset.wasm.mod.Memory().Size())
	}
	return bytes.Clone(bts)
}*/

func (asset *Asset) ptrToBytesDirect(mem uint64) ([]byte, error) {
	if asset.wasm == nil {
		return nil, errors.New("wasm is nil")
	}

	ptr, size := _ptrg(mem)
	bts, ok := asset.wasm.mod.Memory().Read(ptr, size)
	if !ok {
		return nil, fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, asset.wasm.mod.Memory().Size())
	}
	return bts, nil
}

func (asset *Asset) bytesToPtr(src []byte, dst uint64) error {
	if asset.wasm == nil {
		return errors.New("wasm is nil")
	}

	ptr, size := _ptrg(dst)
	n := uint32(len(src))
	if n < size {
		size = n
	}

	//copy string into memory
	if !asset.wasm.mod.Memory().Write(ptr, src) {
		return fmt.Errorf("Memory.Read(%d, %d) out of range of memory size %d", ptr, size, asset.wasm.mod.Memory().Size())
	}
	return nil
}

func (asset *Asset) storage_write(data []byte) (int64, error) {

	//path := asset.app.GetStoragePath(asset.name)
	//err := os.WriteFile(path, data, 0644)
	err := asset.app.root.settings.SetContent(asset.sts_rowid, data)
	if err != nil {
		return -1, err
	}
	return 1, nil
}

func (asset *Asset) _sa_storage_write(jsonStorage uint64) int64 {
	data, err := asset.ptrToBytesDirect(jsonStorage)
	if asset.AddLogErr(err) {
		return -1
	}

	ret, err := asset.storage_write(data)
	asset.AddLogErr(err)
	return ret
}
