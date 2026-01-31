package templates

import (
	"bytes"
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"go.yaml.in/yaml/v4"
)

// FuncMap returns the template FuncMap commonly used in several templates.
func FuncMap() map[string]any {
	return map[string]any{
		"iterate":        FuncIterate,
		"fileContent":    FuncFileContent,
		"secret":         FuncSecret,
		"env":            FuncGetEnv,
		"mustEnv":        FuncMustGetEnv,
		"expandenv":      FuncExpandEnv,
		"split":          FuncStringSplit,
		"splitList":      FuncStringSplitList,
		"join":           FuncElemsJoin,
		"contains":       FuncStringContains,
		"hasPrefix":      FuncStringHasPrefix,
		"hasSuffix":      FuncStringHasSuffix,
		"lower":          strings.ToLower,
		"keys":           FuncKeys,
		"sortAlpha":      FuncSortAlpha,
		"upper":          strings.ToUpper,
		"title":          strings.ToTitle,
		"trim":           strings.TrimSpace,
		"trimAll":        FuncStringTrimAll,
		"trimSuffix":     FuncStringTrimSuffix,
		"trimPrefix":     FuncStringTrimPrefix,
		"replace":        FuncStringReplace,
		"quote":          FuncStringQuote,
		"mquote":         FuncStringQuoteMultiLine(rune(34)),
		"sha1sum":        FuncHashSum(sha1.New),
		"sha256sum":      FuncHashSum(sha256.New),
		"sha512sum":      FuncHashSum(sha512.New),
		"squote":         FuncStringSQuote,
		"msquote":        FuncStringQuoteMultiLine(rune(39)),
		"now":            time.Now,
		"ago":            FuncAgo,
		"toDate":         FuncToDate,
		"mustToDate":     FuncMustToDate,
		"date":           FuncDate,
		"dateInZone":     FuncDateInZone,
		"htmlDate":       FuncHTMLDate,
		"htmlDateInZone": FuncHTMLDateInZone,
		"duration":       FuncDuration,
		"unixEpoch":      FuncUnixEpoch,
		"b64enc":         FuncB64Enc,
		"b64dec":         FuncB64Dec,
		"b32enc":         FuncB32Enc,
		"b32dec":         FuncB32Dec,
		"list":           FuncList,
		"dict":           FuncDict,
		"get":            FuncGet,
		"set":            FuncSet,
		"isAbs":          path.IsAbs,
		"base":           path.Base,
		"dir":            path.Dir,
		"ext":            path.Ext,
		"clean":          path.Clean,
		"osBase":         filepath.Base,
		"osClean":        filepath.Clean,
		"osDir":          filepath.Dir,
		"osExt":          filepath.Ext,
		"osIsAbs":        filepath.IsAbs,
		"deepEqual":      reflect.DeepEqual,
		"typeOf":         FuncTypeOf,
		"typeIs":         FuncTypeIs,
		"typeIsLike":     FuncTypeIsLike,
		"kindOf":         FuncKindOf,
		"kindIs":         FuncKindIs,
		"default":        FuncDefault,
		"empty":          FuncEmpty,
		"indent":         FuncIndent,
		"nindent":        FuncNewlineIndent,
		"mindent":        FuncMultilineIndent,
		"uuidv4":         FuncUUIDv4,
		"urlquery":       url.QueryEscape,
		"urlunquery":     url.QueryUnescape,
		"urlqueryarg":    FuncURLQueryArg,
		"glob":           filepath.Glob,
		"walk":           FuncWalk,

		"fromYaml":     FuncFromYAML,
		"toYaml":       FuncToYAML,
		"toYamlPretty": FuncToYAMLPretty,
		"toYamlCustom": FuncToYAMLCustom,
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
	if key == "$" {
		return key
	}

	if isSecretEnvKey(key) {
		return ""
	}

	value, _ := syscall.Getenv(key)

	return value
}

// FuncMustGetEnv is a special version of os.GetEnv that excludes secret keys and returns an error if it doesn't exist.
func FuncMustGetEnv(key string) (string, error) {
	if key == "$" {
		return key, nil
	}

	value, found := syscall.Getenv(key)

	if !found {
		return "", fmt.Errorf("environment variable '%s' isn't set", key)
	}

	if isSecretEnvKey(key) {
		return "", nil
	}

	return value, nil
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
	var keys []string //nolint:prealloc

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

// FuncStringQuoteMultiLine is a helper function that provides similar functionality
// to FuncStringQuote and FuncStringSQuote, however it skips quoting if the string contains multiple lines.
func FuncStringQuoteMultiLine(char rune) func(in ...any) string {
	return func(in ...any) string {
		out := make([]string, 0, len(in))

		for _, s := range in {
			if s != nil {
				sv := strval(s)

				if strings.Contains(sv, "\n") {
					out = append(out, sv)
				} else {
					out = append(out, fmt.Sprintf("%c%s%c", char, sv, char))
				}
			}
		}

		return strings.Join(out, " ")
	}
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

	return padding + strings.ReplaceAll(value, "\n", "\n"+padding)
}

// FuncNewlineIndent is a helper function that provides similar functionality to the helm nindent func.
func FuncNewlineIndent(indent int, value string) string {
	return "\n" + FuncIndent(indent, value)
}

// FuncMultilineIndent is a helper function that performs YAML multiline intending with a multiline format input such as
// |, |+, |-, >, >+, >-, etc. This is only true if the value has newline characters otherwise it just returns the same
// output as the indent function.
func FuncMultilineIndent(indent int, multiline, value string) string {
	if !strings.Contains(value, "\n") {
		return value
	}

	return multiline + "\n" + FuncIndent(indent, value)
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

// FuncSecret returns the file content stripping the newlines from the end of the content.
func FuncSecret(path string) (data string, err error) {
	if data, err = FuncFileContent(path); err != nil {
		return "", err
	}

	return strings.TrimRight(data, "\n"), nil
}

type WalkInfo struct {
	Path         string
	AbsolutePath string

	os.FileInfo
}

func FuncWalk(root, pattern string, skipDir bool) (infos []WalkInfo, err error) {
	if root == "" {
		return nil, fmt.Errorf("error occurred performing walk: root path cannot be empty")
	}

	var rePattern *regexp.Regexp

	infos = []WalkInfo{}

	if pattern != "" {
		if rePattern, err = regexp.Compile(pattern); err != nil {
			return nil, fmt.Errorf("error occurred compiling walk pattern: %w", err)
		}
	}

	if err = filepath.Walk(root, func(name string, info os.FileInfo, walkErr error) (err error) {
		if walkErr != nil {
			return fmt.Errorf("error occurred walking directory: %w", walkErr)
		}

		walkinfo := WalkInfo{
			Path:     name,
			FileInfo: info,
		}

		if skipDir && walkinfo.IsDir() {
			return nil
		}

		if walkinfo.AbsolutePath, err = filepath.Abs(walkinfo.Path); err != nil {
			return err
		}

		if rePattern != nil && !rePattern.MatchString(walkinfo.AbsolutePath) {
			return nil
		}

		infos = append(infos, walkinfo)

		return nil
	}); err != nil {
		return nil, err
	}

	return infos, nil
}

func FuncFromYAML(yml string) (object map[string]any, err error) {
	object = map[string]any{}

	if err = yaml.Unmarshal([]byte(yml), &object); err != nil {
		return nil, err
	}

	return object, nil
}

func FuncToYAML(object any) (yml string, err error) {
	return FuncToYAMLCustom(object, -1)
}

func FuncToYAMLPretty(object any) (yml string, err error) {
	return FuncToYAMLCustom(object, 2)
}

func FuncToYAMLCustom(object any, indent int) (yml string, err error) {
	var data bytes.Buffer

	encoder := yaml.NewEncoder(&data)

	if indent >= 0 {
		encoder.SetIndent(indent)
	}

	if err = encoder.Encode(object); err != nil {
		return "", err
	}

	return strings.TrimSuffix(data.String(), "\n"), nil
}

func FuncAgo(date any) string {
	return time.Since(convertAnyToTime(date)).Round(time.Second).String()
}

func FuncDate(format string, date any) string {
	return formatTimeWithLocation(format, convertAnyToTime(date), time.Local)
}

func FuncDateInZone(format string, date any, zone string) string {
	var (
		location *time.Location
		err      error
	)
	if location, err = time.LoadLocation(zone); err != nil {
		location = time.UTC
	}

	return formatTimeWithLocation(format, convertAnyToTime(date), location)
}

func FuncHTMLDate(date any) string {
	return formatHTMLTimeWithLocation(convertAnyToTime(date), time.Local)
}

func FuncHTMLDateInZone(date any, zone string) string {
	return FuncDateInZone(time.DateOnly, date, zone)
}

func FuncDuration(sec any) string {
	var n int64

	switch v := sec.(type) {
	case string:
		n, _ = strconv.ParseInt(v, 10, 64)
	case int:
		n = int64(v)
	case int32:
		n = int64(v)
	case int64:
		n = v
	default:
		n = 0
	}

	return (time.Duration(n) * time.Second).String()
}

func FuncToDate(format, date string) time.Time {
	t, _ := time.ParseInLocation(format, date, time.Local)

	return t
}

func FuncMustToDate(format, date string) (time.Time, error) {
	return time.ParseInLocation(format, date, time.Local)
}

func FuncUnixEpoch(date time.Time) string {
	return strconv.FormatInt(date.Unix(), 10)
}

func FuncURLQueryArg(raw, key string) (value string, err error) {
	uri, err := url.ParseRequestURI(raw)
	if err != nil {
		return "", err
	}

	return uri.Query().Get(key), nil
}
