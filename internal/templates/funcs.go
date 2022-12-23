package templates

import (
	"crypto/sha1" //nolint:gosec
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// FuncMap returns the template FuncMap commonly used in several templates.
func FuncMap() map[string]any {
	return map[string]any{
		"iterate":    FuncIterate,
		"env":        FuncGetEnv,
		"expandenv":  FuncExpandEnv,
		"split":      FuncStringsSplit,
		"splitList":  FuncStringSplitList,
		"join":       FuncElemsJoin,
		"contains":   FuncStringContains,
		"hasPrefix":  FuncHasPrefix,
		"hasSuffix":  FuncHasSuffix,
		"lower":      strings.ToLower,
		"upper":      strings.ToUpper,
		"title":      strings.ToTitle,
		"trim":       strings.TrimSpace,
		"trimAll":    FuncTrimAll,
		"trimSuffix": FuncTrimSuffix,
		"trimPrefix": FuncTrimPrefix,
		"replace":    FuncStringReplace,
		"quote":      FuncStringQuote,
		"sha1sum":    FuncHashSum(sha1.New),
		"sha256sum":  FuncHashSum(sha256.New),
		"sha512sum":  FuncHashSum(sha512.New),
		"squote":     FuncStringSQuote,
		"now":        time.Now,
	}
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

// FuncStringReplace is a helper function that provides similar functionality to the helm replace func.
func FuncStringReplace(old, new, s string) string {
	return strings.ReplaceAll(s, old, new)
}

// FuncStringContains is a helper function that provides similar functionality to the helm contains func.
func FuncStringContains(substr string, s string) bool {
	return strings.Contains(s, substr)
}

// FuncHasPrefix is a helper function that provides similar functionality to the helm hasPrefix func.
func FuncHasPrefix(prefix string, s string) bool {
	return strings.HasPrefix(s, prefix)
}

// FuncHasSuffix is a helper function that provides similar functionality to the helm hasSuffix func.
func FuncHasSuffix(suffix string, s string) bool {
	return strings.HasSuffix(s, suffix)
}

// FuncTrimAll is a helper function that provides similar functionality to the helm trimAll func.
func FuncTrimAll(cutset, s string) string {
	return strings.Trim(s, cutset)
}

// FuncTrimSuffix is a helper function that provides similar functionality to the helm trimSuffix func.
func FuncTrimSuffix(suffix, s string) string {
	return strings.TrimSuffix(s, suffix)
}

// FuncTrimPrefix is a helper function that provides similar functionality to the helm trimPrefix func.
func FuncTrimPrefix(prefix, s string) string {
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
			out = append(out, fmt.Sprintf("%q", strval(s)))
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

func strval(v interface{}) string {
	switch v := v.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func strslice(v any) []string {
	switch v := v.(type) {
	case []string:
		return v
	case []interface{}:
		b := make([]string, 0, len(v))

		for _, s := range v {
			if s != nil {
				b = append(b, strval(s))
			}
		}

		return b
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Array, reflect.Slice:
			l := val.Len()
			b := make([]string, 0, l)

			for i := 0; i < l; i++ {
				value := val.Index(i).Interface()
				if value != nil {
					b = append(b, strval(value))
				}
			}

			return b
		default:
			if v == nil {
				return []string{}
			}

			return []string{strval(v)}
		}
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

// FuncStringsSplit is a template function which takes sep and value, splitting the value by the sep into a slice.
func FuncStringsSplit(sep, value string) map[string]string {
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
