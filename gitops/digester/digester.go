/*
Copyright 2024 Adobe. All rights reserved.
This file is licensed to you under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License. You may obtain a copy
of the License at http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR REPRESENTATIONS
OF ANY KIND, either express or implied. See the License for the specific language
governing permissions and limitations under the License.
*/
package digester

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"os"
)

// CalculateDigest calculates the SHA256 digest of a file specified by the given path
func CalculateDigest(path string) string {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return ""
	}

	fi, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer fi.Close()

	h := sha256.New()
	if _, err := io.Copy(h, fi); err != nil {
		log.Fatal(err)
	}

	return hex.EncodeToString(h.Sum(nil))
}

// GetDigest retrieves the digest of a file from a file with the same name but with a ".digest" extension
func GetDigest(path string) string {
	digestPath := path + ".digest"

	if _, err := os.Stat(digestPath); errors.Is(err, os.ErrNotExist) {
		return ""
	}

	digest, err := os.ReadFile(digestPath)
	if err != nil {
		log.Fatal(err)
	}

	return string(digest)
}

// VerifyDigest verifies the integrity of a file by comparing its calculated digest with the stored digest
func VerifyDigest(path string) bool {
	return CalculateDigest(path) == GetDigest(path)
}

// SaveDigest calculates the digest of a file at the given path and saves it to a file with the same name but with a ".digest" extension.
func SaveDigest(path string) {
	digest := CalculateDigest(path)

	digestPath := path + ".digest"

	err := os.WriteFile(digestPath, []byte(digest), 0666)
	if err != nil {
		log.Fatal(err)
	}
}
