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
	"encoding/json"
	"fmt"
	"os"
)

type RS_LResize struct {
	Name  string
	Value float32
}

type RS_LScrollItem struct {
	Hash                   uint64
	ScrollVpos, ScrollHpos int
	Cols_resize            []RS_LResize
	Rows_resize            []RS_LResize
}

type RS_LScroll struct {
	items []*RS_LScrollItem
}

func (scroll *RS_LScroll) Open(path string) error {

	//create ini if not exist
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDONLY, 0644)
	if err != nil {
		return fmt.Errorf("OpenFile() failed: %w", err)
	}
	f.Close()

	js, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("ReadFile(%s) failed: %w", path, err)
	}

	if len(js) > 0 {
		err = json.Unmarshal(js, &scroll.items)
		if err != nil {
			return fmt.Errorf("Unmarshal() failed: %w", err)
		}
	}
	return nil
}

func (scroll *RS_LScroll) Save(path string) error {

	file, err := json.MarshalIndent(&scroll.items, "", "")
	if err != nil {
		return fmt.Errorf("MarshalIndent() failed: %w", err)
	}

	err = os.WriteFile(path, file, 0644)
	if err != nil {
		return fmt.Errorf("WriteFile(%s) failed: %w", path, err)
	}
	return nil
}
