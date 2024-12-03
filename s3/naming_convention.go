package s3

import (
	"errors"
	"fmt"
	"regexp"
	"time"
)

type ObjectType int

const (
    OBJ_TYPE_DUMP ObjectType = iota
    OBJ_TYPE_KEY
)

type ObjectInfo struct {
	Timestamp time.Time
	Type ObjectType
}

type NamingConvention struct {
	Prefix string
	dumpRegex *regexp.Regexp
	keyRegex *regexp.Regexp
	dumpTemplate string
	keyTemplate string
}

func NewNamingConvention(prefix string) NamingConvention {
	return NamingConvention{
		Prefix: prefix,
		dumpRegex: regexp.MustCompile(fmt.Sprintf("^%s-(?P<timestamp>\\d+-\\d+-\\d+T\\d+:\\d+:\\d+Z)\\.dump$", prefix)),
		keyRegex: regexp.MustCompile(fmt.Sprintf("^%s-(?P<timestamp>\\d+-\\d+-\\d+T\\d+:\\d+:\\d+Z)\\.key$", prefix)),
		dumpTemplate: fmt.Sprintf("%s-%%s.dump", prefix),
		keyTemplate: fmt.Sprintf("%s-%%s.key", prefix),
	}
}

func (conv *NamingConvention) GetObjectNames(timestamp time.Time) (string, string) {
	timeStr := timestamp.Format(time.RFC3339)

	return fmt.Sprintf(conv.dumpTemplate, timeStr),
		fmt.Sprintf(conv.keyTemplate, timeStr)
}

func (conv *NamingConvention) GetObjectInfo(objName string) (ObjectInfo, error) {
	if conv.dumpRegex.MatchString(objName) {
		match := conv.dumpRegex.FindStringSubmatch(objName)

		t, parseErr := time.Parse(time.RFC3339, match[1])
		if parseErr != nil {
			return ObjectInfo{}, errors.New(fmt.Sprintf("Timestamp '%s' in object '%s' does not parse properly", match[1], objName))
		}

		return ObjectInfo{
			Timestamp: t,
			Type: OBJ_TYPE_DUMP,
		}, nil
	}

	if conv.keyRegex.MatchString(objName) {
		match := conv.keyRegex.FindStringSubmatch(objName)

		t, parseErr := time.Parse(time.RFC3339, match[1])
		if parseErr != nil {
			return ObjectInfo{}, errors.New(fmt.Sprintf("Timestamp '%s' in object '%s' does not parse properly", match[1], objName))
		}

		return ObjectInfo{
			Timestamp: t,
			Type: OBJ_TYPE_KEY,
		}, nil
	}

	return ObjectInfo{}, errors.New(fmt.Sprintf("Object name '%s' does not match the expected object name format", objName))
}