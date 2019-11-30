// Copyright (c) Derek Author(s) 2017. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package auth

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
)

// CheckMAC verifies hash checksum
func CheckMAC(message, messageMAC, key []byte) bool {
	mac := hmac.New(sha1.New, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)

	return hmac.Equal(messageMAC, expectedMAC)
}

// ValidateHMAC validate a digest from Github via xHubSignature
func ValidateHMAC(secret string, bytesIn []byte, xHubSignature string) error {
	if len(xHubSignature) > 5 {
		messageMAC := xHubSignature[5:] // first few chars are: sha1=
		messageMACBuf, _ := hex.DecodeString(messageMAC)
		res := CheckMAC(bytesIn, []byte(messageMACBuf), []byte(secret))
		if !res {
			return fmt.Errorf("invalid message digest or secret")
		}
	}
	return nil
}
