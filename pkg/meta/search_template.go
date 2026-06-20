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

package meta

import "time"

type SearchTemplate struct {
	Name      string                 `json:"name"`
	Template  string                 `json:"template"`
	Params    map[string]interface{} `json:"params,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

type SearchTemplateRequest struct {
	Template string                 `json:"template"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

type SearchTemplateRenderRequest struct {
	ID     string                 `json:"id"`
	Params map[string]interface{} `json:"params,omitempty"`
}
