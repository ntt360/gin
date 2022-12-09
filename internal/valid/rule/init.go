package rule

import (
	"net/textproto"
	"reflect"
	"strings"
	"time"
)

// Func accepts a Field interface for all validation needs. The return
// value should be true when validation succeeds.
type Func func(fl Field) bool

var (
	TimeType = reflect.TypeOf(time.Time{})

	ValidatorFunc = map[string]Func{
		"required":         hasValue,
		"required_if":      requiredIf,
		"len":              hasLengthOf,
		"min":              hasMinOf,
		"max":              hasMaxOf,
		"eq":               isEq,
		"ne":               isNe,
		"lt":               isLt,
		"lte":              isLte,
		"gt":               isGt,
		"gte":              isGte,
		"eqfield":          isEqField,
		"eqcsfield":        isEqCrossStructField,
		"necsfield":        isNeCrossStructField,
		"gtcsfield":        isGtCrossStructField,
		"gtecsfield":       isGteCrossStructField,
		"ltcsfield":        isLtCrossStructField,
		"ltecsfield":       isLteCrossStructField,
		"nefield":          isNeField,
		"gtefield":         isGteField,
		"gtfield":          isGtField,
		"ltefield":         isLteField,
		"ltfield":          isLtField,
		"fieldcontains":    fieldContains,
		"fieldexcludes":    fieldExcludes,
		"alpha":            isAlpha,
		"alphanum":         isAlphanum,
		"alphaunicode":     isAlphaUnicode,
		"alphanumunicode":  isAlphanumUnicode,
		"boolean":          isBoolean,
		"numeric":          isNumeric,
		"number":           isNumber,
		"email":            isEmail,
		"url":              isURL,
		"uri":              isURI,
		"urn_rfc2141":      isUrnRFC2141, // RFC 2141
		"file":             isFile,
		"base64":           isBase64,
		"base64url":        isBase64URL,
		"contains":         contains,
		"containsany":      containsAny,
		"containsrune":     containsRune,
		"excludes":         excludes,
		"excludesall":      excludesAll,
		"excludesrune":     excludesRune,
		"startswith":       startsWith,
		"endswith":         endsWith,
		"startsnotwith":    startsNotWith,
		"endsnotwith":      endsNotWith,
		"md5":              isMD5,
		"sha256":           isSHA256,
		"ipv4":             isIPv4,
		"ipv6":             isIPv6,
		"ip":               isIP,
		"tcp4_addr":        isTCP4AddrResolvable,
		"tcp6_addr":        isTCP6AddrResolvable,
		"tcp_addr":         isTCPAddrResolvable,
		"udp4_addr":        isUDP4AddrResolvable,
		"udp6_addr":        isUDP6AddrResolvable,
		"udp_addr":         isUDPAddrResolvable,
		"ip4_addr":         isIP4AddrResolvable,
		"ip6_addr":         isIP6AddrResolvable,
		"ip_addr":          isIPAddrResolvable,
		"unix_addr":        isUnixAddrResolvable,
		"mac":              isMAC,
		"hostname":         isHostnameRFC952,  // RFC 952
		"hostname_rfc1123": isHostnameRFC1123, // RFC 1123
		"unique":           isUnique,
		"oneof":            isOneOf,
		"html":             isHTML,
		"html_encoded":     isHTMLEncoded,
		"url_encoded":      isURLEncoded,
		"json":             isJSON,
		"jwt":              isJWT,
		"hostname_port":    isHostnamePort,
		"lowercase":        isLowercase,
		"uppercase":        isUppercase,
		"datetime":         isDatetime,
		"timezone":         isTimeZone,

		// new ext validator
		"idcard": isIdCard,
		"mobile": isMobile,
		"regex":  isRegexMatch,
	}
)

func Head(str, sep string) (head string, tail string) {
	idx := strings.Index(str, sep)
	if idx < 0 {
		return str, ""
	}

	return str[:idx], str[idx+len(sep):]
}

func Param(tag reflect.StructTag, param string) string {
	tagValue, _ := Head(tag.Get(param), ",")

	if param == "header" {
		tagValue = textproto.CanonicalMIMEHeaderKey(tagValue)
	}

	return tagValue
}

func DefaultVal(tag reflect.StructTag, param string) (string, bool) {
	_, opts := Head(tag.Get(param), ",")
	var opt string
	for len(opts) > 0 {
		opt, opts = Head(opts, ",")

		if k, v := Head(opt, "="); k == "default" {
			return v, true
		}
	}

	return "", false
}
