package security

import "crypto/hmac"
import "crypto/sha256"

func LookupHash(secret, value string) []byte {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(value))
	return h.Sum(nil)
}
