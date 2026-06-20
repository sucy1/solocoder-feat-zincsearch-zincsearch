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

package core

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/zincsearch/zincsearch/pkg/errors"
	"github.com/zincsearch/zincsearch/pkg/meta"
	"github.com/zincsearch/zincsearch/pkg/metadata"
	"github.com/zincsearch/zincsearch/pkg/zutils/json"
)

var templateParamRegex = regexp.MustCompile(`\{\{\s*(\w+)(?::([^}]*))?\s*\}\}`)

func ListSearchTemplates() ([]*meta.SearchTemplate, error) {
	templates, err := metadata.SearchTemplate.List(0, 0)
	if err != nil {
		return nil, err
	}
	if templates == nil {
		templates = make([]*meta.SearchTemplate, 0)
	}
	return templates, nil
}

func GetSearchTemplate(name string) (*meta.SearchTemplate, bool, error) {
	if name == "" {
		return nil, false, nil
	}
	tpl, err := metadata.SearchTemplate.Get(name)
	if err != nil {
		if err == errors.ErrKeyNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return tpl, true, nil
}

func CreateSearchTemplate(name string, template string, defaultParams map[string]interface{}) error {
	if name == "" {
		return errors.New(errors.ErrorTypeParsingException, "template name is required")
	}
	if template == "" {
		return errors.New(errors.ErrorTypeParsingException, "template body is required")
	}

	if err := ValidateSearchTemplate(template); err != nil {
		return err
	}

	now := time.Now()
	tpl := meta.SearchTemplate{
		Name:      name,
		Template:  template,
		Params:    defaultParams,
		CreatedAt: now,
		UpdatedAt: now,
	}
	return metadata.SearchTemplate.Set(name, tpl)
}

func UpdateSearchTemplate(name string, template string, defaultParams map[string]interface{}) error {
	if name == "" {
		return errors.New(errors.ErrorTypeParsingException, "template name is required")
	}

	existing, exists, err := GetSearchTemplate(name)
	if err != nil {
		return err
	}

	if err := ValidateSearchTemplate(template); err != nil {
		return err
	}

	now := time.Now()
	createdAt := now
	if exists {
		createdAt = existing.CreatedAt
	}
	tpl := meta.SearchTemplate{
		Name:      name,
		Template:  template,
		Params:    defaultParams,
		CreatedAt: createdAt,
		UpdatedAt: now,
	}
	return metadata.SearchTemplate.Set(name, tpl)
}

func DeleteSearchTemplate(name string) error {
	if name == "" {
		return errors.New(errors.ErrorTypeParsingException, "template name is required")
	}
	return metadata.SearchTemplate.Delete(name)
}

func RenderSearchTemplate(template string, params map[string]interface{}) (string, error) {
	if template == "" {
		return "", errors.New(errors.ErrorTypeParsingException, "template is empty")
	}

	result := templateParamRegex.ReplaceAllStringFunc(template, func(match string) string {
		submatches := templateParamRegex.FindStringSubmatch(match)
		if len(submatches) < 2 {
			return match
		}
		paramName := strings.TrimSpace(submatches[1])
		defaultValue := ""
		if len(submatches) >= 3 {
			defaultValue = strings.TrimSpace(submatches[2])
		}

		value, ok := params[paramName]
		if !ok {
			if defaultValue != "" {
				return defaultValue
			}
			return match
		}

		switch v := value.(type) {
		case string:
			return v
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return match
			}
			return string(b)
		}
	})

	remainingMatches := templateParamRegex.FindAllString(result, -1)
	if len(remainingMatches) > 0 {
		missingParams := make([]string, 0, len(remainingMatches))
		seen := make(map[string]bool)
		for _, m := range remainingMatches {
			submatches := templateParamRegex.FindStringSubmatch(m)
			if len(submatches) >= 2 {
				paramName := strings.TrimSpace(submatches[1])
				if !seen[paramName] {
					missingParams = append(missingParams, paramName)
					seen[paramName] = true
				}
			}
		}
		return "", errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("template parameters missing: %s", strings.Join(missingParams, ", ")))
	}

	return result, nil
}

func RenderSearchTemplateWithDefaults(name string, params map[string]interface{}) (string, error) {
	tpl, exists, err := GetSearchTemplate(name)
	if err != nil {
		return "", err
	}
	if !exists {
		return "", errors.New(errors.ErrorTypeParsingException, fmt.Sprintf("search template [%s] not found", name))
	}

	mergedParams := make(map[string]interface{})
	for k, v := range tpl.Params {
		mergedParams[k] = v
	}
	for k, v := range params {
		mergedParams[k] = v
	}

	return RenderSearchTemplate(tpl.Template, mergedParams)
}

func ValidateSearchTemplate(template string) error {
	if template == "" {
		return errors.New(errors.ErrorTypeParsingException, "template is empty")
	}

	if !strings.Contains(template, "{") && !strings.Contains(template, "}") {
		return nil
	}

	openBraces := strings.Count(template, "{{")
	closeBraces := strings.Count(template, "}}")
	if openBraces != closeBraces {
		return errors.New(errors.ErrorTypeParsingException, "template syntax error: mismatched {{ and }}")
	}

	matches := templateParamRegex.FindAllStringSubmatch(template, -1)
	for _, match := range matches {
		if len(match) < 2 || match[1] == "" {
			return errors.New(errors.ErrorTypeParsingException, "template syntax error: empty parameter name")
		}
		paramName := strings.TrimSpace(match[1])
		if paramName == "" {
			return errors.New(errors.ErrorTypeParsingException, "template syntax error: empty parameter name")
		}
	}

	return nil
}
