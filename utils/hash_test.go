package utils

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestHashes(t *testing.T) {
	testCases := []struct{
		isMatched bool
		data string
		inputHash string
		expectedError string
	} {
		{
			isMatched: true,
			data:      "one two three",
			inputHash: "md5=5e4fe0155703dde467f3ab234e6f966f",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "md5=5e4fe0155703dde467f3ab234e6f966f1",
		},
		{
			isMatched: true,
			data:      "one two three",
			inputHash: "sha1=a10600b129253b1aaaa860778bef2043ee40c715",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "sha1=a10600b129253b1aaaa860778bef2043ee40c716",
		},
		{
			isMatched: true,
			data:      "one two three",
			inputHash: "sha224=edbb09e64f0172798322de92f4b89be23170d9e41744c5e6d17a647e",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "sha224=edbb09e64f0172798322de92f4b89be23170d9e41744c5e6d17a647e1",
		},
		{
			isMatched: true,
			data:      "one two three",
			inputHash: "sha256=6899ee404683a14e8c2a03149860df25d67d34d9cd4dae7350cbe91e4b3976be",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "sha256=6899ee404683a14e8c2a03149860df25d67d34d9cd4dae7350cbe91e4b3976be1",
		},
		{
			isMatched: true,
			data:      "one two three",
			inputHash: "sha384=e1748a3a4f4a475298e282d47eb88fd8cebfb38d49b6d74bf67220c4f55e5a6020de9b8937a630ab15e9e3739c80e252",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "sha384=3e1748a3a4f4a475298e282d47eb88fd8cebfb38d49b6d74bf67220c4f55e5a6020de9b8937a630ab15e9e3739c80e252",
		},
		{
			isMatched: true,
			data:      "one two three",
			inputHash: "sha512=c11c7e22fa98c41c21324a1da6d56eaa3d558bae4979ec9a4f74d56b62cf698c5a7bf0b2f12f18f886ef2346ad6793cb5161ee7d41bc9d5cba6767a76cc1357b",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "sha512=1c11c7e22fa98c41c21324a1da6d56eaa3d558bae4979ec9a4f74d56b62cf698c5a7bf0b2f12f18f886ef2346ad6793cb5161ee7d41bc9d5cba6767a76cc1357b",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "md5=",
			expectedError: "invalid hash string 'md5='",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "",
			expectedError: "invalid hash string ''",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "ddd",
			expectedError: "invalid hash string 'ddd'",
		},
		{
			isMatched: false,
			data:      "one two three",
			inputHash: "ddd=2222",
			expectedError: "unknown hash algorithm 'ddd' in 'ddd=2222'",
		},
	}

	for _, testCase := range testCases {
		err := ioutil.WriteFile("testFile.txt", []byte(testCase.data), 0644)
		assert.NoError(t, err)
		if err != nil {
			return
		}

		actualMatch, _, err := HashEquals(testCase.inputHash, "testFile.txt")
		if testCase.expectedError != "" {
			assert.EqualError(t, err, testCase.expectedError)
			continue
		}

		assert.NoError(t, err)
		if err != nil {
			continue
		}

		assert.Equal(t, testCase.isMatched, actualMatch, "hash '%s' should match with the actual hash of '%s'", testCase.inputHash, testCase.data)
	}

	err := os.Remove("testFile.txt")
	assert.NoError(t, err)
}
