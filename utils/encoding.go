package utils

import (
	"fmt"
	"golang.org/x/text/encoding/charmap"
	"io/ioutil"
	"os"
	"strings"
)

func Encode(encodingName, contentsUtf8 string) ([]byte, error) {
	cm, err := DetectCharMap(encodingName)
	if err != nil {
		return []byte{}, err
	}

	enc := cm.NewEncoder()
	out, err := enc.String(contentsUtf8)

	return []byte(out), err
}

func Decode(encodingName string, data []byte) (string, error) {
	cm, err := DetectCharMap(encodingName)
	if err != nil {
		return "", err
	}

	enc := cm.NewDecoder()
	out, err := enc.Bytes(data)

	return string(out), err
}

func DetectCharMap(encodingName string) (*charmap.Charmap, error) {
	var cm *charmap.Charmap
	var err error

	switch strings.ToLower(encodingName) {
	case "codepage037":
		cm = charmap.CodePage037
	case "codepage1047":
		cm = charmap.CodePage1047
	case "codepage1140":
		cm = charmap.CodePage1140
	case "codepage437":
		cm = charmap.CodePage437
	case "codepage850":
		cm = charmap.CodePage850
	case "codepage852":
		cm = charmap.CodePage852
	case "codepage855":
		cm = charmap.CodePage855
	case "codepage858":
		cm = charmap.CodePage858
	case "codepage860":
		cm = charmap.CodePage860
	case "codepage862":
		cm = charmap.CodePage862
	case "codepage863":
		cm = charmap.CodePage863
	case "codepage865":
		cm = charmap.CodePage865
	case "codepage866":
		cm = charmap.CodePage866
	case "iso8859_1":
		cm = charmap.ISO8859_1
	case "iso8859_10":
		cm = charmap.ISO8859_10
	case "iso8859_13":
		cm = charmap.ISO8859_13
	case "iso8859_14":
		cm = charmap.ISO8859_14
	case "iso8859_15":
		cm = charmap.ISO8859_15
	case "iso8859_16":
		cm = charmap.ISO8859_16
	case "iso8859_2":
		cm = charmap.ISO8859_2
	case "iso8859_3":
		cm = charmap.ISO8859_3
	case "iso8859_4":
		cm = charmap.ISO8859_4
	case "iso8859_5":
		cm = charmap.ISO8859_5
	case "iso8859_6":
		cm = charmap.ISO8859_6
	case "iso8859_7":
		cm = charmap.ISO8859_7
	case "iso8859_8":
		cm = charmap.ISO8859_8
	case "iso8859_9":
		cm = charmap.ISO8859_9
	case "koi8r":
		cm = charmap.KOI8R
	case "koi8u":
		cm = charmap.KOI8U
	case "macintosh":
		cm = charmap.Macintosh
	case "macintoshcyrillic":
		cm = charmap.MacintoshCyrillic
	case "windows1250":
		cm = charmap.Windows1250
	case "windows1251":
		cm = charmap.Windows1251
	case "windows1252":
		cm = charmap.Windows1252
	case "windows1253":
		cm = charmap.Windows1253
	case "windows1254":
		cm = charmap.Windows1254
	case "windows1255":
		cm = charmap.Windows1255
	case "windows1256":
		cm = charmap.Windows1256
	case "windows1257":
		cm = charmap.Windows1257
	case "windows1258":
		cm = charmap.Windows1258
	case "windows874":
		cm = charmap.Windows874
	default:
		err = fmt.Errorf("unknown encoding: '%s'", encodingName)
	}

	return cm, err
}

func WriteEncodedFile(encodingName, contentsUtf8, fileName string, perm os.FileMode) error {
	encodedData, err := Encode(encodingName, contentsUtf8)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(fileName, encodedData, perm)
}
