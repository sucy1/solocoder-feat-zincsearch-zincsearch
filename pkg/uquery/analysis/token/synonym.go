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

package token

import (
	"bytes"
	"strings"
	"sync"

	"github.com/blugelabs/bluge/analysis"

	"github.com/zincsearch/zincsearch/pkg/errors"
	"github.com/zincsearch/zincsearch/pkg/zutils"
)

type SynonymTokenFilter struct {
	expand       bool
	synonyms     map[string][]string
	originalCase bool
	lock         sync.RWMutex
}

func NewSynonymTokenFilter(options interface{}) (analysis.TokenFilter, error) {
	synonymsList, err := zutils.GetStringSliceFromMap(options, "synonyms")
	if err != nil {
		return nil, errors.New(errors.ErrorTypeParsingException, "[token_filter] synonym option [synonyms] should be an array of string")
	}

	expand, _ := zutils.GetBoolFromMap(options, "expand")
	originalCase, _ := zutils.GetBoolFromMap(options, "original_case")

	synonyms := make(map[string][]string)
	for _, rule := range synonymsList {
		parseSynonymRule(rule, synonyms, expand)
	}

	return &SynonymTokenFilter{
		expand:       expand,
		synonyms:     synonyms,
		originalCase: originalCase,
	}, nil
}

func parseSynonymRule(rule string, synonyms map[string][]string, expand bool) {
	rule = strings.TrimSpace(rule)
	if rule == "" {
		return
	}

	if strings.Contains(rule, "=>") {
		parts := strings.SplitN(rule, "=>", 2)
		if len(parts) != 2 {
			return
		}
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		if left == "" || right == "" {
			return
		}

		leftWords := splitAndTrim(left)
		rightWords := splitAndTrim(right)

		for _, lw := range leftWords {
			if lw == "" {
				continue
			}
			key := strings.ToLower(lw)
			synonyms[key] = appendUnique(synonyms[key], rightWords)
		}
	} else {
		words := splitAndTrim(rule)
		if len(words) < 2 {
			return
		}

		for i, w := range words {
			if w == "" {
				continue
			}
			key := strings.ToLower(w)
			for j, sw := range words {
				if i != j && sw != "" {
					synonyms[key] = appendUnique(synonyms[key], []string{sw})
				}
			}
		}
	}
}

func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

func appendUnique(slice []string, items []string) []string {
	existing := make(map[string]bool)
	for _, s := range slice {
		existing[s] = true
	}
	for _, item := range items {
		if !existing[item] {
			slice = append(slice, item)
			existing[item] = true
		}
	}
	return slice
}

func (f *SynonymTokenFilter) Filter(input analysis.TokenStream) analysis.TokenStream {
	f.lock.RLock()
	defer f.lock.RUnlock()

	output := make(analysis.TokenStream, 0, len(input))
	for _, tok := range input {
		output = append(output, tok)

		term := string(tok.Term)
		key := strings.ToLower(term)

		if synonyms, ok := f.synonyms[key]; ok && len(synonyms) > 0 {
			for _, syn := range synonyms {
				synTerm := syn
				if !f.originalCase {
					synTerm = strings.ToLower(synTerm)
				}
				if synTerm == term || synTerm == key {
					continue
				}
				newTok := &analysis.Token{
					Term:         []byte(synTerm),
					PositionIncr: 0,
					Start:        tok.Start,
					End:          tok.Start + len(synTerm),
					Type:         tok.Type,
					KeyWord:      tok.KeyWord,
				}
				output = append(output, newTok)
			}
		}
	}
	return output
}

func (f *SynonymTokenFilter) Reload(synonymsList []string) {
	f.lock.Lock()
	defer f.lock.Unlock()

	synonyms := make(map[string][]string)
	for _, rule := range synonymsList {
		parseSynonymRule(rule, synonyms, f.expand)
	}
	f.synonyms = synonyms
}

func (f *SynonymTokenFilter) GetSynonyms() map[string][]string {
	f.lock.RLock()
	defer f.lock.RUnlock()

	result := make(map[string][]string, len(f.synonyms))
	for k, v := range f.synonyms {
		result[k] = make([]string, len(v))
		copy(result[k], v)
	}
	return result
}

var globalSynonymFilters = make(map[string]*SynonymTokenFilter)
var globalSynonymLock sync.RWMutex

func RegisterSynonymFilter(name string, filter *SynonymTokenFilter) {
	globalSynonymLock.Lock()
	defer globalSynonymLock.Unlock()
	globalSynonymFilters[name] = filter
}

func ReloadSynonymFilter(name string, synonymsList []string) bool {
	globalSynonymLock.RLock()
	defer globalSynonymLock.RUnlock()
	if filter, ok := globalSynonymFilters[name]; ok {
		filter.Reload(synonymsList)
		return true
	}
	return false
}

func GetAllSynonymFilters() map[string]map[string][]string {
	globalSynonymLock.RLock()
	defer globalSynonymLock.RUnlock()

	result := make(map[string]map[string][]string, len(globalSynonymFilters))
	for name, filter := range globalSynonymFilters {
		result[name] = filter.GetSynonyms()
	}
	return result
}

func compareBytes(a, b []byte) bool {
	return bytes.Equal(a, b)
}
