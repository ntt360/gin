package rule

import "regexp"

const (
	alphaRegexString               = "^[a-zA-Z]+$"
	alphaNumericRegexString        = "^[a-zA-Z0-9]+$"
	alphaUnicodeRegexString        = "^[\\p{L}]+$"
	alphaUnicodeNumericRegexString = "^[\\p{L}\\p{N}]+$"
	NumericRegexString             = "^[-+]?[0-9]+(?:\\.[0-9]+)?$"
	numberRegexString              = "^[0-9]+$"
	emailRegexString               = "^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$"
	base64RegexString              = "^(?:[A-Za-z0-9+\\/]{4})*(?:[A-Za-z0-9+\\/]{2}==|[A-Za-z0-9+\\/]{3}=|[A-Za-z0-9+\\/]{4})$"
	base64URLRegexString           = "^(?:[A-Za-z0-9-_]{4})*(?:[A-Za-z0-9-_]{2}==|[A-Za-z0-9-_]{3}=|[A-Za-z0-9-_]{4})$"
	md5RegexString                 = "^[0-9a-f]{32}$"
	sha256RegexString              = "^[0-9a-f]{64}$"
	hostnameRegexStringRFC952      = `^[a-zA-Z]([a-zA-Z0-9\-]+[\.]?)*[a-zA-Z0-9]$`                                   // https://tools.ietf.org/html/rfc952
	hostnameRegexStringRFC1123     = `^([a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62}){1}(\.[a-zA-Z0-9]{1}[a-zA-Z0-9-]{0,62})*?$` // accepts hostname starting with a digit https://tools.ietf.org/html/rfc1123
	uRLEncodedRegexString          = `^(?:[^%]|%[0-9A-Fa-f]{2})*$`
	hTMLEncodedRegexString         = `&#[x]?([0-9a-fA-F]{2})|(&gt)|(&lt)|(&quot)|(&amp)+[;]?`
	hTMLRegexString                = `<[/]?([a-zA-Z]+).*?>`
	jWTRegexString                 = "^[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]+\\.[A-Za-z0-9-_]*$"
	splitParamsRegexString         = `'[^']*'|\S+`
	mobile                         = "^(?:\\+?86)?1(?:3\\d{3}|5[^4\\D]\\d{2}|8\\d{3}|7(?:[35678]\\d{2}|4(?:0\\d|1[0-2]|9\\d))|9[189]\\d{2}|66\\d{2})\\d{6}$"
)

var (
	alphaRegex               = regexp.MustCompile(alphaRegexString)
	alphaNumericRegex        = regexp.MustCompile(alphaNumericRegexString)
	alphaUnicodeRegex        = regexp.MustCompile(alphaUnicodeRegexString)
	alphaUnicodeNumericRegex = regexp.MustCompile(alphaUnicodeNumericRegexString)
	numberRegex              = regexp.MustCompile(numberRegexString)
	emailRegex               = regexp.MustCompile(emailRegexString)
	base64Regex              = regexp.MustCompile(base64RegexString)
	base64URLRegex           = regexp.MustCompile(base64URLRegexString)
	md5Regex                 = regexp.MustCompile(md5RegexString)
	sha256Regex              = regexp.MustCompile(sha256RegexString)
	hostnameRegexRFC952      = regexp.MustCompile(hostnameRegexStringRFC952)
	hostnameRegexRFC1123     = regexp.MustCompile(hostnameRegexStringRFC1123)
	uRLEncodedRegex          = regexp.MustCompile(uRLEncodedRegexString)
	hTMLEncodedRegex         = regexp.MustCompile(hTMLEncodedRegexString)
	hTMLRegex                = regexp.MustCompile(hTMLRegexString)
	jWTRegex                 = regexp.MustCompile(jWTRegexString)
	splitParamsRegex         = regexp.MustCompile(splitParamsRegexString)
	mobileRegex              = regexp.MustCompile(mobile)
)
