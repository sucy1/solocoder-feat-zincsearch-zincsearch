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

package index

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/zincsearch/zincsearch/pkg/meta"
	zinctoken "github.com/zincsearch/zincsearch/pkg/uquery/analysis/token"
	"github.com/zincsearch/zincsearch/pkg/zutils"
)

type SynonymReloadRequest struct {
	Name     string   `json:"name"`
	Synonyms []string `json:"synonyms"`
}

// @Id ListSynonyms
// @Summary List all synonym filters
// @security BasicAuth
// @Tags    Index
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /es/_synonyms [get]
func ListSynonyms(c *gin.Context) {
	synonyms := zinctoken.GetAllSynonymFilters()
	zutils.GinRenderJSON(c, http.StatusOK, gin.H{"synonyms": synonyms})
}

// @Id ReloadSynonym
// @Summary Reload synonym filter with new rules (hot reload)
// @security BasicAuth
// @Tags    Index
// @Accept  json
// @Produce json
// @Param   name path string true "Synonym filter name"
// @Param   request body SynonymReloadRequest true "Synonym rules"
// @Success 200 {object} meta.HTTPResponse
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/_synonyms/{name} [post]
func ReloadSynonym(c *gin.Context) {
	name := c.Param("target")

	var req SynonymReloadRequest
	if err := zutils.GinBindJSON(c, &req); err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	if name == "" {
		name = req.Name
	}
	if name == "" {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: "synonym filter name is required"})
		return
	}

	ok := zinctoken.ReloadSynonymFilter(name, req.Synonyms)
	if !ok {
		zutils.GinRenderJSON(c, http.StatusNotFound, meta.HTTPResponseError{Error: "synonym filter " + name + " not found"})
		return
	}

	zutils.GinRenderJSON(c, http.StatusOK, meta.HTTPResponse{Message: "synonym filter " + name + " reloaded successfully"})
}

// @Id GetSynonym
// @Summary Get a synonym filter by name
// @security BasicAuth
// @Tags    Index
// @Produce json
// @Param   name path string true "Synonym filter name"
// @Success 200 {object} map[string]interface{}
// @Failure 404 {object} meta.HTTPResponseError
// @Router /es/_synonyms/{name} [get]
func GetSynonym(c *gin.Context) {
	name := c.Param("target")
	if name == "" {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: "synonym filter name is required"})
		return
	}

	all := zinctoken.GetAllSynonymFilters()
	synonyms, ok := all[name]
	if !ok {
		zutils.GinRenderJSON(c, http.StatusNotFound, meta.HTTPResponseError{Error: "synonym filter " + name + " not found"})
		return
	}

	zutils.GinRenderJSON(c, http.StatusOK, gin.H{"name": name, "synonyms": synonyms})
}
