// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"sort"
	"strings"
	"unicode"
)

const conformanceModulePrefix = "oidcc-"

// ConformanceSubtestNames returns the PascalCase subtest name of each module, index aligned with modules. A module
// which a plan lists more than once has its variant appended so the names within a plan stay unique, which is what
// makes go test able to address a single one with -run.
func ConformanceSubtestNames(modules []ConformancePlanModule) (names []string) {
	counts := make(map[string]int, len(modules))

	for _, module := range modules {
		counts[module.TestModule]++
	}

	names = make([]string, len(modules))

	for i, module := range modules {
		name := conformancePascalCase(strings.TrimPrefix(module.TestModule, conformanceModulePrefix))

		if counts[module.TestModule] > 1 {
			name += conformanceVariantSuffix(module.Variant)
		}

		names[i] = name
	}

	return names
}

func conformanceVariantSuffix(variant map[string]string) (suffix string) {
	keys := make([]string, 0, len(variant))

	for key := range variant {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	for _, key := range keys {
		suffix += conformancePascalCase(variant[key])
	}

	return suffix
}

func conformancePascalCase(value string) string {
	builder := &strings.Builder{}

	for _, part := range strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	}) {
		runes := []rune(strings.ToLower(part))
		runes[0] = unicode.ToUpper(runes[0])

		builder.WriteString(string(runes))
	}

	return builder.String()
}
