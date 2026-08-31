package utils

import (
	"encoding/base64"
	"log"
	"net/url"
	"reflect"
	"regexp"
	"regexp/syntax"
	"strconv"
	"strings"

	"github.com/gfunc/subconvergo/config"
	"github.com/gfunc/subconvergo/proxy/core"
	"github.com/gfunc/subconvergo/proxy/impl"
	"github.com/metacubex/mihomo/adapter"
)

func ToMihomoProxy(pObj core.ParsableProxy) (core.ParsableProxy, error) {
	return ToMihomoProxyWithSetting(pObj, &config.ProxySetting{})
}

func ToMihomoProxyFromClash(pObj core.ParsableProxy, config map[string]interface{}) (core.ParsableProxy, error) {
	// Sanitize subscription-controlled fields before they can reach any
	// line-oriented output.
	SanitizeProxy(pObj)
	for k, v := range config {
		config[k] = sanitizeAny(v)
	}
	// If it's already a MihomoProxy, just update the options
	if mp, ok := pObj.(*impl.MihomoProxy); ok {
		mp.Options = config
		return mp, nil
	}

	mihomoProxy, err := adapter.ParseProxy(config)
	if err != nil {
		log.Printf("[ToMihomoProxyFromClash] Failed to parse proxy with adapter: %v", err)
		return &impl.MihomoProxy{
			ProxyInterface: pObj,
			Options:        config,
		}, nil
	}

	return &impl.MihomoProxy{
		ProxyInterface: pObj,
		Clash:          mihomoProxy,
		Options:        config,
	}, nil
}

func ToMihomoProxyWithSetting(pObj core.ParsableProxy, config *config.ProxySetting) (core.ParsableProxy, error) {
	// Sanitize subscription-controlled fields before they can reach any
	// line-oriented output.
	SanitizeProxy(pObj)
	if _, ok := pObj.(*impl.MihomoProxy); ok {
		return pObj, nil
	}
	if oObj, ok := pObj.(core.ClashConvertableMixin); ok {

		option, err := oObj.ToClashConfig(config)
		if err != nil {
			log.Printf("[toMihomoProxy] Failed to convert proxy to Clash config: %v", err)
			return pObj, nil
		}

		mihomoProxy, err := adapter.ParseProxy(option)
		if err != nil {
			log.Printf("[toMihomoProxy] Converted proxy: %+v to Mihomo format: %+v, err: %v", pObj, mihomoProxy, err)
			return pObj, nil
		} else {
			return &impl.MihomoProxy{
				ProxyInterface: pObj,
				Clash:          mihomoProxy,
				Options:        option,
			}, nil
		}
	}
	return pObj, nil
}

func GetStringField(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case string:
			return val
		case float64:
			return strconv.FormatFloat(val, 'f', -1, 64)
		case int:
			return strconv.Itoa(val)
		}
	}
	return ""
}

// Helper functions (duplicated for now, should be in a shared utils package)
func UrlDecode(s string) string {
	decoded, err := url.QueryUnescape(s)
	if err != nil {
		return s
	}
	return decoded
}

func UrlSafeBase64Decode(s string) string {
	s = strings.ReplaceAll(s, "-", "+")
	s = strings.ReplaceAll(s, "_", "/")
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	decoded, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(s)
		if err != nil {
			return s
		}
	}
	return string(decoded)
}

func ParsePluginOpts(opts string) map[string]interface{} {
	result := make(map[string]interface{})
	pairs := strings.Split(opts, ";")
	for _, pair := range pairs {
		if pair == "" {
			continue
		}
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			result[kv[0]] = UrlDecode(kv[1])
		} else {
			result[kv[0]] = "true"
		}
	}
	return result
}

func GetIntField(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case int:
			return val
		case string:
			i, _ := strconv.Atoi(val)
			return i
		}
	}
	return 0
}

func GetBoolField(m map[string]interface{}, key string) bool {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case bool:
			return val
		case string:
			return val == "true"
		}
	}
	return false
}

func GetBoolPtrField(m map[string]interface{}, key string) *bool {
	if v, ok := m[key]; ok {
		var b bool
		switch val := v.(type) {
		case bool:
			b = val
		case string:
			if val == "true" || val == "1" {
				b = true
			} else if val == "false" || val == "0" {
				b = false
			} else {
				return nil
			}
		default:
			return nil
		}
		return &b
	}
	return nil
}

func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return ""
	}
}

func ToInt(v interface{}) int {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return int(val)
	case int:
		return val
	case string:
		i, _ := strconv.Atoi(val)
		return i
	default:
		return 0
	}
}

// Security hardening helpers.

// SanitizeScalarField strips ASCII control characters (0x00-0x1F and 0x7F),
// including CR, LF and NUL, from a subscription-controlled scalar field.
//
// Line-oriented outputs (Surge, Loon, Quantumult X, single-link lists)
// interpolate these fields with raw string formatting; a single \n in a
// remark/password/host used to inject whole INI sections (e.g. a crafted
// ss:// fragment `#legit%0A[rewrite_local]%0A...`). Stripping (rather than
// rejecting the node) keeps legitimate nodes working; all printable
// characters — commas, `=`, quotes, unicode — pass through unchanged.
func SanitizeScalarField(s string) string {
	if strings.IndexFunc(s, func(r rune) bool { return r < 0x20 || r == 0x7f }) == -1 {
		return s
	}
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

// SanitizeProxy sanitizes every string field of a parsed proxy in place:
// remarks, servers, credentials, group, underlying-proxy, plugin
// names/options, SNI/host/path, IDs, string slices (ALPN/DNS) and
// url.Values params. For *impl.MihomoProxy the wrapped proxy and the raw
// Options map are sanitized as well.
//
// It is called from the ToMihomoProxy* funnel every parser/proxy Parse*
// method returns through, so all protocols and subscription formats are
// covered at the parse/normalization boundary.
func SanitizeProxy(pObj core.ParsableProxy) {
	if pObj == nil {
		return
	}
	if mp, ok := pObj.(*impl.MihomoProxy); ok {
		if mp.ProxyInterface != nil {
			sanitizeProxyValue(reflect.ValueOf(mp.ProxyInterface))
		}
		if mp.Options != nil {
			for k, v := range mp.Options {
				mp.Options[k] = sanitizeAny(v)
			}
		}
		return
	}
	sanitizeProxyValue(reflect.ValueOf(pObj))
}

func sanitizeProxyValue(v reflect.Value) {
	if v.Kind() == reflect.Ptr && !v.IsNil() && v.Elem().Kind() == reflect.Struct {
		sanitizeStructFields(v.Elem())
	}
}

func sanitizeStructFields(v reflect.Value) {
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if t.Field(i).PkgPath != "" { // unexported
			continue
		}
		f := v.Field(i)
		switch f.Kind() {
		case reflect.String:
			if f.CanSet() {
				f.SetString(SanitizeScalarField(f.String()))
			}
		case reflect.Struct:
			sanitizeStructFields(f)
		case reflect.Ptr:
			if !f.IsNil() && f.Elem().Kind() == reflect.Struct {
				sanitizeStructFields(f.Elem())
			}
		case reflect.Slice:
			for j := 0; j < f.Len(); j++ {
				e := f.Index(j)
				if e.Kind() == reflect.String && e.CanSet() {
					e.SetString(SanitizeScalarField(e.String()))
				}
			}
		case reflect.Map:
			sanitizeMapValue(f)
		}
	}
}

func sanitizeMapValue(f reflect.Value) {
	if f.IsNil() || f.Type().Key().Kind() != reflect.String {
		return
	}
	et := f.Type().Elem()
	switch {
	case et.Kind() == reflect.Interface: // map[string]interface{}
		for _, k := range f.MapKeys() {
			f.SetMapIndex(k, reflect.ValueOf(sanitizeAny(f.MapIndex(k).Interface())))
		}
	case et.Kind() == reflect.String: // map[string]string
		for _, k := range f.MapKeys() {
			f.SetMapIndex(k, reflect.ValueOf(SanitizeScalarField(f.MapIndex(k).String())))
		}
	case et.Kind() == reflect.Slice && et.Elem().Kind() == reflect.String: // map[string][]string, e.g. url.Values
		for _, k := range f.MapKeys() {
			vals := f.MapIndex(k)
			out := reflect.MakeSlice(et, vals.Len(), vals.Len())
			for j := 0; j < vals.Len(); j++ {
				out.Index(j).SetString(SanitizeScalarField(vals.Index(j).String()))
			}
			f.SetMapIndex(k, out)
		}
	}
}

// sanitizeAny returns the value with all contained strings sanitized,
// recursing into nested maps and slices. Non-string scalars are untouched.
func sanitizeAny(v interface{}) interface{} {
	switch t := v.(type) {
	case string:
		return SanitizeScalarField(t)
	case map[string]interface{}:
		for k, val := range t {
			t[k] = sanitizeAny(val)
		}
		return t
	case []interface{}:
		for i := range t {
			t[i] = sanitizeAny(t[i])
		}
		return t
	case []string:
		for i := range t {
			t[i] = SanitizeScalarField(t[i])
		}
		return t
	}
	return v
}

// Source proxy-group regex filter guard.

const (
	// MaxGroupFilterLength caps the length of a source-provided proxy-group
	// regex filter. Longer filters are stripped (not truncated: a truncated
	// regex is a different regex).
	MaxGroupFilterLength = 256
	// MaxSourceProxyGroups caps how many proxy-groups a single subscription
	// may contribute. Excess groups are dropped at parse time.
	MaxSourceProxyGroups = 100
)

// IsSafeGroupFilter reports whether a source-provided proxy-group filter is
// safe to honor: it must compile under Go's RE2 regexp engine and contain no
// catastrophic-backtracking shapes — nested quantifiers (`^(a+)+$`, `(a*)+`)
// or a quantifier over an alternation with duplicate/empty-matchable
// branches (`(a|a)*`) — so it stays linear-time even if it is ever handed to
// a backtracking engine downstream (e.g. mihomo's regexp2, which has no
// default match timeout).
func IsSafeGroupFilter(filter string) bool {
	if filter == "" || len(filter) > MaxGroupFilterLength {
		return false
	}
	if _, err := regexp.Compile(filter); err != nil {
		return false
	}
	re, err := syntax.Parse(filter, syntax.Perl)
	if err != nil {
		return false
	}
	return !hasUnsafeQuantifier(re) && !hasDuplicateBranchQuantifier(filter)
}

func isQuantifierOp(op syntax.Op) bool {
	switch op {
	case syntax.OpStar, syntax.OpPlus, syntax.OpQuest, syntax.OpRepeat:
		return true
	}
	return false
}

func unwrapCaptures(re *syntax.Regexp) *syntax.Regexp {
	for re.Op == syntax.OpCapture {
		re = re.Sub[0]
	}
	return re
}

// hasUnsafeQuantifier walks the parsed pattern looking for quantifiers that
// can cause exponential backtracking in backtracking engines.
func hasUnsafeQuantifier(re *syntax.Regexp) bool {
	if isQuantifierOp(re.Op) && quantifierSubtreeUnsafe(unwrapCaptures(re.Sub[0])) {
		return true
	}
	for _, sub := range re.Sub {
		if hasUnsafeQuantifier(sub) {
			return true
		}
	}
	return false
}

// quantifierSubtreeUnsafe reports whether a quantified subtree contains
// another quantifier (nested quantifier: (a+)+, (a*)+) or an alternation
// with an empty-matchable branch ((a|ab)* — RE2's parser factors it to
// a(?:|b), so the empty branch check catches it after factoring).
func quantifierSubtreeUnsafe(re *syntax.Regexp) bool {
	re = unwrapCaptures(re)
	if isQuantifierOp(re.Op) {
		return true
	}
	if re.Op == syntax.OpAlternate {
		for _, branch := range re.Sub {
			if matchesEmpty(branch) {
				return true
			}
		}
	}
	for _, sub := range re.Sub {
		if quantifierSubtreeUnsafe(sub) {
			return true
		}
	}
	return false
}

// matchesEmpty reports whether the sub-pattern can match the empty string.
func matchesEmpty(re *syntax.Regexp) bool {
	re = unwrapCaptures(re)
	switch re.Op {
	case syntax.OpEmptyMatch:
		return true
	case syntax.OpStar, syntax.OpQuest:
		return true
	case syntax.OpRepeat:
		return re.Min == 0 || matchesEmpty(re.Sub[0])
	case syntax.OpConcat:
		for _, sub := range re.Sub {
			if !matchesEmpty(sub) {
				return false
			}
		}
		return true
	}
	return false
}

// hasDuplicateBranchQuantifier scans the raw pattern for a quantified group
// whose top-level alternation has duplicate or empty branches — `(a|a)*`.
// This must run on the raw text because regexp/syntax's parser factors such
// alternations away (`(a|a)` parses to `(a)`) before the AST check sees them.
func hasDuplicateBranchQuantifier(p string) bool {
	for i := 0; i < len(p); i++ {
		switch p[i] {
		case '\\':
			i++
		case '[':
			i++
			for i < len(p) {
				if p[i] == '\\' {
					i++
				} else if p[i] == ']' {
					break
				}
				i++
			}
		case '(':
			end := matchParen(p, i)
			if end == -1 {
				return false // unbalanced; regexp.Compile already rejects this
			}
			if isQuantifierAt(p, end+1) {
				branches := splitTopLevelAlternation(groupBodyText(p, i, end))
				if len(branches) > 1 {
					seen := make(map[string]bool, len(branches))
					for _, b := range branches {
						if b == "" || seen[b] {
							return true
						}
						seen[b] = true
					}
				}
			}
			// Deliberately do not skip ahead: nested groups are scanned too.
		}
	}
	return false
}

// matchParen returns the index of the ')' matching the '(' at open, or -1.
func matchParen(p string, open int) int {
	depth := 0
	for i := open; i < len(p); i++ {
		switch p[i] {
		case '\\':
			i++
		case '[':
			i++
			for i < len(p) {
				if p[i] == '\\' {
					i++
				} else if p[i] == ']' {
					break
				}
				i++
			}
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// isQuantifierAt reports whether p[j] starts a quantifier (* + ? {m,n}).
func isQuantifierAt(p string, j int) bool {
	if j >= len(p) {
		return false
	}
	switch p[j] {
	case '*', '+', '?':
		return true
	case '{':
		end := strings.IndexByte(p[j:], '}')
		return end != -1
	}
	return false
}

// groupBodyText strips the group opener syntax ((?:...) (?flags:...)
// (?P<name>...)) from the body of the group spanning p[open:close].
func groupBodyText(p string, open, close int) string {
	body := p[open+1 : close]
	if !strings.HasPrefix(body, "?") {
		return body
	}
	if idx := strings.IndexByte(body, ':'); idx != -1 {
		return body[idx+1:]
	}
	if strings.HasPrefix(body, "?P<") {
		if idx := strings.IndexByte(body, '>'); idx != -1 {
			return body[idx+1:]
		}
	}
	return "" // (?flags) group: carries no alternation
}

// splitTopLevelAlternation splits a group body on top-level '|' (ignoring
// escaped chars, char classes and nested groups).
func splitTopLevelAlternation(body string) []string {
	var branches []string
	depth := 0
	inClass := false
	start := 0
	for i := 0; i < len(body); i++ {
		c := body[i]
		if c == '\\' {
			i++
			continue
		}
		if inClass {
			if c == ']' {
				inClass = false
			}
			continue
		}
		switch c {
		case '[':
			inClass = true
		case '(':
			depth++
		case ')':
			depth--
		case '|':
			if depth == 0 {
				branches = append(branches, body[start:i])
				start = i + 1
			}
		}
	}
	return append(branches, body[start:])
}
