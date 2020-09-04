package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"regexp"

	log "github.com/sirupsen/logrus"
)

func HashEquals(hashStr, filePath string) (hashEquals bool, actualCache string, err error) {
	log.Debugf("will check if file %s matches hash %s", filePath, hashStr)

	fileExists, err := FileExists(filePath)
	if err != nil {
		log.Warnf("file exists check failure: %v", err)
	}

	if !fileExists {
		log.Debugf("file '%s' doesn't exist, hash check is skipped for it", filePath)
		return false, "", nil
	}

	expectedHashAlgoName, expectedHashSum, err := ParseHashAlgoAndSum(hashStr)
	if err != nil {
		return false, "", err
	}

	actualHashSum, err := HashSum(expectedHashAlgoName, filePath)
	if err != nil {
		return false, "", err
	}

	isMatched := expectedHashSum == actualHashSum
	if isMatched {
		log.Debugf("file hash at '%s' is matched", filePath)
	} else {
		log.Debugf("file hash at '%s' didn't match", filePath)
	}

	return isMatched, fmt.Sprintf("%s=%s", expectedHashAlgoName, actualHashSum), nil
}

func HashSum(hashAlgoName, filePath string) (hashSum string, err error) {
	hashAlgo, err := ExtractHashAlgo(hashAlgoName)
	if err != nil {
		return "", err
	}

	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}

	defer CloseResourceSecure(filePath, f)

	if _, err := io.Copy(hashAlgo, f); err != nil {
		return "", err
	}

	hashSum = fmt.Sprintf("%x", hashAlgo.Sum(nil))

	return
}

func ParseHashAlgoAndSum(hashStr string) (algoName, sum string, err error) {
	const expectedRegexParts = 3
	reg := regexp.MustCompile(`^(\w*)=(.+)$`)
	regParts := reg.FindStringSubmatch(hashStr)

	if len(regParts) != expectedRegexParts {
		return "", "", fmt.Errorf("invalid hash string '%s'", hashStr)
	}

	return regParts[1], regParts[2], nil
}

func ExtractHashAlgo(hashAlgoName string) (hChecker hash.Hash, err error) {
	switch hashAlgoName {
	case "sha512":
		return sha512.New(), nil
	case "sha384":
		return sha512.New384(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha224":
		return sha256.New224(), nil
	case "sha1":
		return sha1.New(), nil
	case "md5":
		return md5.New(), nil
	default:
		return nil, fmt.Errorf("unknown hash algorithm '%s'", hashAlgoName)
	}
}
