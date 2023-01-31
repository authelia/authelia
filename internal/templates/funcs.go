package templates

import (
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// FuncMap returns the template FuncMap commonly used in several templates.
func FuncMap() map[string]any {
	return map[string]any{
		"iterate":     FuncIterate,
		"fileContent": FuncFileContent,
		"env":         FuncGetEnv,
		"expandenv":   FuncExpandEnv,
		"split":       FuncStringSplit,
		"splitList":   FuncStringSplitList,
		"join":        FuncElemsJoin,
		"contains":    FuncStringContains,
		"hasPrefix":   FuncStringHasPrefix,
		"hasSuffix":   FuncStringHasSuffix,
		"lower":       strings.ToLower,
		"keys":        FuncKeys,
		"sortAlpha":   FuncSortAlpha,
		"upper":       strings.ToUpper,
		"title":       strings.ToTitle,
		"trim":        strings.TrimSpace,
		"trimAll":     FuncStringTrimAll,
		"trimSuffix":  FuncStringTrimSuffix,
		"trimPrefix":  FuncStringTrimPrefix,
		"replace":     FuncStringReplace,
		"quote":       FuncStringQuote,
		"sha1sum":     FuncHashSum(sha1.New),
		"sha256sum":   FuncHashSum(sha256.New),
		"sha512sum":   FuncHashSum(sha512.New),
		"squote":      FuncStringSQuote,
		"now":         time.Now,
		"b64enc":      FuncB64Enc,
		"b64dec":      FuncB64Dec,
		"b32enc":      FuncB32Enc,
		"b32dec":      FuncB32Dec,
		"list":        FuncList,
		"dict":        FuncDict,
		"get":         FuncGet,
		"set":         FuncSet,
		"isAbs":       path.IsAbs,
		"base":        path.Base,
		"dir":         path.Dir,
		"ext":         path.Ext,
		"clean":       path.Clean,
		"osBase":      filepath.Base,
		"osClean":     filepath.Clean,
		"osDir":       filepath.Dir,
		"osExt":       filepath.Ext,
		"osIsAbs":     filepath.IsAbs,
		"deepEqual":   reflect.DeepEqual,
		"typeOf":      FuncTypeOf,
		"typeIs":      FuncTypeIs,
		"typeIsLike":  FuncTypeIsLike,
		"kindOf":      FuncKindOf,
		"kindIs":      FuncKindIs,
		"default":     FuncDefault,
		"empty":       FuncEmpty,
		"indent":      FuncIndent,
		"nindent":     FuncNewlineIndent,
		"uuidv4":      FuncUUIDv4,
	}
}

// FuncB64Enc is a helper function that provides similar functionality to the helm b64enc func.
func FuncB64Enc(input string) string {
	return base64.StdEncoding.EncodeToString([]byte(input))
}

// FuncB64Dec is a helper function that provides similar functionality to the helm b64dec func.
func FuncB64Dec(input string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FuncB32Enc is a helper function that provides similar functionality to the helm b32enc func.
func FuncB32Enc(input string) string {
	return base32.StdEncoding.EncodeToString([]byte(input))
}

// FuncB32Dec is a helper function that provides similar functionality to the helm b32dec func.
func FuncB32Dec(input string) (string, error) {
	data, err := base32.StdEncoding.DecodeString(input)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// FuncExpandEnv is a special version of os.ExpandEnv that excludes secret keys.
func FuncExpandEnv(s string) string {
	return os.Expand(s, FuncGetEnv)
}

// FuncGetEnv is a special version of os.GetEnv that excludes secret keys.
func FuncGetEnv(key string) string {
	if isSecretEnvKey(key) {
		return ""
	}

	return os.Getenv(key)
}

// FuncHashSum is a helper function that provides similar functionality to helm sum funcs.
func FuncHashSum(new func() hash.Hash) func(data string) string {
	hasher := new()

	return func(data string) string {
		sum := hasher.Sum([]byte(data))

		return hex.EncodeToString(sum)
	}
}

// FuncKeys is a helper function that provides similar functionality to the helm keys func.
func FuncKeys(maps ...map[string]any) []string {
	var keys []string

	for _, m := range maps {
		for k := range m {
			keys = append(keys, k)
		}
	}

	return keys
}

// FuncSortAlpha is a helper function that provides similar functionality to the helm sortAlpha func.
func FuncSortAlpha(slice any) []string {
	kind := reflect.Indirect(reflect.ValueOf(slice)).Kind()

	switch kind {
	case reflect.Slice, reflect.Array:
		unsorted := strslice(slice)
		sorted := sort.StringSlice(unsorted)
		sorted.Sort()

		return sorted
	}

	return []string{strval(slice)}
}

// FuncStringReplace is a helper function that provides similar functionality to the helm replace func.
func FuncStringReplace(old, new, s string) string {
	return strings.ReplaceAll(s, old, new)
}

// FuncStringContains is a helper function that provides similar functionality to the helm contains func.
func FuncStringContains(substr string, s string) bool {
	return strings.Contains(s, substr)
}

// FuncStringHasPrefix is a helper function that provides similar functionality to the helm hasPrefix func.
func FuncStringHasPrefix(prefix string, s string) bool {
	return strings.HasPrefix(s, prefix)
}

// FuncStringHasSuffix is a helper function that provides similar functionality to the helm hasSuffix func.
func FuncStringHasSuffix(suffix string, s string) bool {
	return strings.HasSuffix(s, suffix)
}

// FuncStringTrimAll is a helper function that provides similar functionality to the helm trimAll func.
func FuncStringTrimAll(cutset, s string) string {
	return strings.Trim(s, cutset)
}

// FuncStringTrimSuffix is a helper function that provides similar functionality to the helm trimSuffix func.
func FuncStringTrimSuffix(suffix, s string) string {
	return strings.TrimSuffix(s, suffix)
}

// FuncStringTrimPrefix is a helper function that provides similar functionality to the helm trimPrefix func.
func FuncStringTrimPrefix(prefix, s string) string {
	return strings.TrimPrefix(s, prefix)
}

// FuncElemsJoin is a helper function that provides similar functionality to the helm join func.
func FuncElemsJoin(sep string, elems any) string {
	return strings.Join(strslice(elems), sep)
}

// FuncStringSQuote is a helper function that provides similar functionality to the helm squote func.
func FuncStringSQuote(in ...any) string {
	out := make([]string, 0, len(in))

	for _, s := range in {
		if s != nil {
			out = append(out, fmt.Sprintf("'%s'", strval(s)))
		}
	}

	return strings.Join(out, " ")
}

// FuncStringQuote is a helper function that provides similar functionality to the helm quote func.
func FuncStringQuote(in ...any) string {
	out := make([]string, 0, len(in))

	for _, s := range in {
		if s != nil {
			out = append(out, fmt.Sprintf("%q", strval(s)))
		}
	}

	return strings.Join(out, " ")
}

// FuncIterate is a template function which takes a single uint returning a slice of units from 0 up to that number.
func FuncIterate(count *uint) (out []uint) {
	var i uint

	for i = 0; i < (*count); i++ {
		out = append(out, i)
	}

	return
}

// FuncStringSplit is a template function which takes sep and value, splitting the value by the sep into a slice.
func FuncStringSplit(sep, value string) map[string]string {
	parts := strings.Split(value, sep)
	res := make(map[string]string, len(parts))

	for i, v := range parts {
		res["_"+strconv.Itoa(i)] = v
	}

	return res
}

// FuncStringSplitList is a special split func that reverses the inputs to match helm templates.
func FuncStringSplitList(sep, s string) []string {
	return strings.Split(s, sep)
}

// FuncStringJoinX takes a list of string elements, joins them by the sep string, before every int n characters are
// written it writes string p. This is useful for line breaks mostly.
func FuncStringJoinX(elems []string, sep string, n int, p string) string {
	buf := strings.Builder{}

	c := 0
	e := len(elems) - 1

	for i := 0; i <= e; i++ {
		if c+len(elems[i])+1 > n {
			c = 0

			buf.WriteString(p)
		}

		c += len(elems[i]) + 1

		buf.WriteString(elems[i])

		if i < e {
			buf.WriteString(sep)
		}
	}

	return buf.String()
}

// FuncTypeIs is a helper function that provides similar functionality to the helm typeIs func.
func FuncTypeIs(is string, v any) bool {
	return is == FuncTypeOf(v)
}

// FuncTypeIsLike is a helper function that provides similar functionality to the helm typeIsLike func.
func FuncTypeIsLike(is string, v any) bool {
	t := FuncTypeOf(v)

	return is == t || "*"+is == t
}

// FuncTypeOf is a helper function that provides similar functionality to the helm typeOf func.
func FuncTypeOf(v any) string {
	return reflect.ValueOf(v).Type().String()
}

// FuncKindIs is a helper function that provides similar functionality to the helm kindIs func.
func FuncKindIs(is string, v any) bool {
	return is == FuncKindOf(v)
}

// FuncKindOf is a helper function that provides similar functionality to the helm kindOf func.
func FuncKindOf(v any) string {
	return reflect.ValueOf(v).Kind().String()
}

// FuncList is a helper function that provides similar functionality to the helm list func.
func FuncList(items ...any) []any {
	return items
}

// FuncDict is a helper function that provides similar functionality to the helm dict func.
func FuncDict(pairs ...any) map[string]any {
	m := map[string]any{}
	p := len(pairs)

	for i := 0; i < p; i += 2 {
		key := strval(pairs[i])

		if i+1 >= p {
			m[key] = ""

			continue
		}

		m[key] = pairs[i+1]
	}

	return m
}

// FuncGet is a helper function that provides similar functionality to the helm get func.
func FuncGet(m map[string]any, key string) any {
	if val, ok := m[key]; ok {
		return val
	}

	return ""
}

// FuncSet is a helper function that provides similar functionality to the helm set func.
func FuncSet(m map[string]any, key string, value any) map[string]any {
	m[key] = value

	return m
}

// FuncDefault is a helper function that provides similar functionality to the helm default func.
func FuncDefault(d any, vals ...any) any {
	if FuncEmpty(vals) || FuncEmpty(vals[0]) {
		return d
	}

	return vals[0]
}

// FuncEmpty is a helper function that provides similar functionality to the helm empty func.
func FuncEmpty(v any) bool {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return true
	}

	switch rv.Kind() {
	default:
		return rv.IsNil()
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
		return rv.Len() == 0
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Struct:
		return false
	}
}

// FuncIndent is a helper function that provides similar functionality to the helm indent func.
func FuncIndent(indent int, value string) string {
	padding := strings.Repeat(" ", indent)

	return padding + strings.Replace(value, "\n", "\n"+padding, -1)
}

// FuncNewlineIndent is a helper function that provides similar functionality to the helm nindent func.
func FuncNewlineIndent(indent int, value string) string {
	return "\n" + FuncIndent(indent, value)
}

// FuncUUIDv4 is a helper function that provides similar functionality to the helm uuidv4 func.
func FuncUUIDv4() string {
	return uuid.New().String()
}

// FuncFileContent returns the file content.
func FuncFileContent(path string) (data string, err error) {
	var raw []byte

	if raw, err = os.ReadFile(path); err != nil {
		return "", err
	}

	return string(raw), nil
}
