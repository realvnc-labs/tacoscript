package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	log "github.com/sirupsen/logrus"
	"hash"
	"io"
	"os"
	"regexp"
)

func HashEquals(hashStr, filePath string) (bool, error) {
	log.Debugf("will check if file %s matches hash %s", filePath, hashStr)

	fileExists, err := FileExists(filePath)
	if err != nil {
		return false, err
	}
	if !fileExists {
		log.Debugf("file %s doesn't exist, so it should be created", filePath)
		return false, nil
	}

	f, err := os.Open(filePath)
	if err != nil {
		return false, err
	}

	defer CloseResourceSecure(filePath, f)

	hashAlgo, expectedHashStr, err := ExtractHashAlgo(hashStr)
	if err != nil {
		return false, err
	}

	if _, err := io.Copy(hashAlgo, f); err != nil {
		return false, err
	}

	isMatched := expectedHashStr == fmt.Sprintf("%x", hashAlgo.Sum(nil))
	if isMatched {
		log.Debugf("file hash at '%s' is matched", filePath)
	} else {
		log.Debugf("file hash at '%s' didn't match", filePath)
	}

	return isMatched, nil
}

func ExtractHashAlgo(hashStr string) (hash.Hash, string, error) {
	reg := regexp.MustCompile(`^(\w*)=(.+)$`)
	regParts := reg.FindStringSubmatch(hashStr)

	if len(regParts) != 3 {
		return nil, "", fmt.Errorf("invalid hash string '%s'", hashStr)
	}

	switch regParts[1] {
	case "sha512":
		return sha512.New(), regParts[2], nil
	case "sha384":
		return sha512.New384(), regParts[2], nil
	case "sha256":
		return sha256.New(), regParts[2], nil
	case "sha224":
		return sha256.New224(), regParts[2], nil
	case "sha1":
		return sha1.New(), regParts[2], nil
	case "md5":
		return md5.New(), regParts[2], nil
	default:
		return nil, "", fmt.Errorf("unknown hash algorithm '%s' in '%s'", regParts[1], hashStr)
	}
}
