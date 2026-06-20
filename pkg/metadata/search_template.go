/* Copyright 2022 Zinc Labs Inc. and Contributors
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package metadata

import (
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

type searchTemplate struct{}

var SearchTemplate = new(searchTemplate)

func (t *searchTemplate) List(offset, limit int) ([]*meta.SearchTemplate, error) {
	data, err := db.List(t.key(""), offset, limit)
	if err != nil {
		return nil, err
	}
	templates := make([]*meta.SearchTemplate, 0, len(data))
	for _, d := range data {
		tpl := new(meta.SearchTemplate)
		err = json.Unmarshal(d, tpl)
		if err != nil {
			return nil, err
		}
		templates = append(templates, tpl)
	}
	return templates, nil
}

func (t *searchTemplate) Get(id string) (*meta.SearchTemplate, error) {
	data, err := db.Get(t.key(id))
	if err != nil {
		return nil, err
	}
	tpl := new(meta.SearchTemplate)
	err = json.Unmarshal(data, tpl)
	return tpl, err
}

func (t *searchTemplate) Set(id string, val meta.SearchTemplate) error {
	data, err := json.Marshal(val)
	if err != nil {
		return err
	}
	return db.Set(t.key(id), data)
}

func (t *searchTemplate) Delete(id string) error {
	return db.Delete(t.key(id))
}

func (t *searchTemplate) key(id string) string {
	return "/search_template/" + id
}
