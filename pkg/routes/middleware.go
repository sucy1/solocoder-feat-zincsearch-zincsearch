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

package routes

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/zincsearch/zincsearch/pkg/auth"
	"github.com/zincsearch/zincsearch/pkg/core"
)

func AuthMiddleware(permission string) func(c *gin.Context) {
	auth.AddPermission(permission)
	return func(c *gin.Context) {
		// Get the Basic Authentication credentials
		user, password, hasAuth := c.Request.BasicAuth()
		if hasAuth {
			if u, ok := auth.VerifyCredentials(user, password); ok {
				if auth.VerifyRoleHasPermission(u.Role, permission) {
					c.Next()
				} else {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "No permission:" + permission})
					return
				}
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"auth": "Invalid credentials"})
				return
			}
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"auth": "Missing credentials"})
			return
		}
	}
}

func ESMiddleware(c *gin.Context) {
	// Some es clients will check header("X-elastic-product") == "Elasticsearch".
	// If not, it will not work, and show "The client noticed that the server is not Elasticsearch and we do not support this unknown product."
	c.Header("X-elastic-product", "Elasticsearch")
}

func IndexAliasMiddleware(c *gin.Context) {
	target := ""
	ix := 0

	for i, entry := range c.Params {
		if entry.Key == "target" {
			target = entry.Value
			ix = i
			break
		}
	}

	if target == "" {
		c.Next()
		return
	}

	isWrite := isWriteRequest(c)

	if isWrite {
		writeIndex, ok := core.ZINC_INDEX_ALIAS_LIST.GetWriteIndexForAlias(target)
		if ok {
			c.Params[ix].Value = writeIndex
		}
	} else {
		indexes, ok := core.ZINC_INDEX_ALIAS_LIST.GetIndexesForAlias(target)
		if !ok {
			c.Next()
			return
		}
		newTarget := strings.Join(indexes, ",")
		if newTarget != "" {
			c.Params[ix].Value = newTarget
		}
	}
	c.Next()
}

func isWriteRequest(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	if method == http.MethodPost || method == http.MethodPut || method == http.MethodDelete {
		if strings.Contains(path, "/_doc") ||
			strings.Contains(path, "/_create") ||
			strings.Contains(path, "/_update") ||
			strings.Contains(path, "/_bulk") ||
			strings.Contains(path, "/_bulkv2") ||
			strings.Contains(path, "/_multi") ||
			strings.Contains(path, "/_delete_by_query") {
			return true
		}
	}
	return false
}
