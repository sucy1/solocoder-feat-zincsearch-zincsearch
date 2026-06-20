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

package search

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zincsearch/zincsearch/pkg/core"
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/zutils"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

type CreateSearchTemplateRequest struct {
	Template string                 `json:"template"`
	Params   map[string]interface{} `json:"params,omitempty"`
}

// @Id ListSearchTemplates
// @Summary List all search templates
// @security BasicAuth
// @Tags    Search
// @Produce json
// @Success 200 {object} []meta.SearchTemplate
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/_search/template [get]
func ListSearchTemplates(c *gin.Context) {
	templates, err := core.ListSearchTemplates()
	if err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}
	zutils.GinRenderJSON(c, http.StatusOK, templates)
}

// @Id GetSearchTemplate
// @Summary Get a search template by name
// @security BasicAuth
// @Tags    Search
// @Produce json
// @Param   name path string true "Template name"
// @Success 200 {object} meta.SearchTemplate
// @Failure 404 {object} meta.HTTPResponseError
// @Router /es/_search/template/{name} [get]
func GetSearchTemplate(c *gin.Context) {
	name := c.Param("target")
	if name == "" {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: "template name is required"})
		return
	}
	template, exists, err := core.GetSearchTemplate(name)
	if err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}
	if !exists {
		zutils.GinRenderJSON(c, http.StatusNotFound, meta.HTTPResponseError{Error: "search template " + name + " not found"})
		return
	}
	zutils.GinRenderJSON(c, http.StatusOK, template)
}

// @Id CreateSearchTemplate
// @Summary Create or update a search template
// @security BasicAuth
// @Tags    Search
// @Accept  json
// @Produce json
// @Param   name path string true "Template name"
// @Param   request body CreateSearchTemplateRequest true "Template data"
// @Success 200 {object} meta.HTTPResponse
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/_search/template/{name} [post]
func CreateSearchTemplate(c *gin.Context) {
	name := c.Param("target")

	var req CreateSearchTemplateRequest
	if err := zutils.GinBindJSON(c, &req); err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	if name == "" {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: "template name is required"})
		return
	}

	err := core.CreateSearchTemplate(name, req.Template, req.Params)
	if err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	zutils.GinRenderJSON(c, http.StatusOK, meta.HTTPResponse{Message: "search template " + name + " created successfully"})
}

// @Id UpdateSearchTemplate
// @Summary Update a search template
// @security BasicAuth
// @Tags    Search
// @Accept  json
// @Produce json
// @Param   name path string true "Template name"
// @Param   request body CreateSearchTemplateRequest true "Template data"
// @Success 200 {object} meta.HTTPResponse
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/_search/template/{name} [put]
func UpdateSearchTemplate(c *gin.Context) {
	name := c.Param("target")

	var req CreateSearchTemplateRequest
	if err := zutils.GinBindJSON(c, &req); err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	if name == "" {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: "template name is required"})
		return
	}

	err := core.UpdateSearchTemplate(name, req.Template, req.Params)
	if err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	zutils.GinRenderJSON(c, http.StatusOK, meta.HTTPResponse{Message: "search template " + name + " updated successfully"})
}

// @Id DeleteSearchTemplate
// @Summary Delete a search template
// @security BasicAuth
// @Tags    Search
// @Produce json
// @Param   name path string true "Template name"
// @Success 200 {object} meta.HTTPResponse
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/_search/template/{name} [delete]
func DeleteSearchTemplate(c *gin.Context) {
	name := c.Param("target")
	err := core.DeleteSearchTemplate(name)
	if err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}
	zutils.GinRenderJSON(c, http.StatusOK, meta.HTTPResponse{Message: "search template " + name + " deleted successfully"})
}

// @Id RenderSearchTemplate
// @Summary Render and execute a search template
// @security BasicAuth
// @Tags    Search
// @Accept  json
// @Produce json
// @Param   index path string true "Index name"
// @Param   request body meta.SearchTemplateRenderRequest true "Template render request"
// @Success 200 {object} meta.SearchResponse
// @Failure 400 {object} meta.HTTPResponseError
// @Router /es/{index}/_search/template [post]
func RenderAndSearchTemplate(c *gin.Context) {
	indexName := c.Param("target")

	var body map[string]interface{}
	if err := zutils.GinBindJSON(c, &body); err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	var templateName string
	var params map[string]interface{}

	if v, ok := body["id"]; ok {
		templateName, _ = v.(string)
	}
	if v, ok := body["params"]; ok {
		if pm, ok := v.(map[string]interface{}); ok {
			params = pm
		}
	}

	if templateName == "" {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: "template id is required"})
		return
	}

	rendered, err := core.RenderSearchTemplateWithDefaults(templateName, params)
	if err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	query := &meta.ZincQuery{Size: 10}
	if err := json.Unmarshal([]byte(rendered), query); err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: "rendered template is not valid JSON: " + err.Error()})
		return
	}

	resp, err := searchIndex(strings.Split(indexName, ","), query)
	if err != nil {
		zutils.GinRenderJSON(c, http.StatusBadRequest, meta.HTTPResponseError{Error: err.Error()})
		return
	}

	zutils.GinRenderJSON(c, http.StatusOK, resp)
}
