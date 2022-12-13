package crypto

import "encoding/base64"

// Base64Encode base64 encode
func Base64Encode(data []byte) []byte {
	coded := make([]byte, base64.StdEncoding.EncodedLen(len(data)))
	base64.StdEncoding.Encode(coded, data)

	return coded
}

// Base64Decode base64 decode
func Base64Decode(coded []byte) []byte {
	data := make([]byte, base64.StdEncoding.DecodedLen(len(coded)))
	n, _ := base64.StdEncoding.Decode(data, coded)

	return data[:n]
}

// Base64EncodeString encode from string
func Base64EncodeString(data string) string {
	return string(Base64Encode([]byte(data)))
}

// Base64DecodeString decode from string
func Base64DecodeString(data string) string {
	return string(Base64Decode([]byte(data)))
}
