package std

import (
	"strings"

	"github.com/PondWader/kit/pkg/lang/values"
)

var ParseVersion = values.Of(parseVersion)

func parseVersion(version values.Value) (values.Value, *values.Error) {
	str, ok := version.ToString()
	if !ok {
		return values.Nil, values.FmtTypeError("parse_version", values.KindString)
	}

	return values.Of(values.ObjectFromStruct(parsedVersion{raw: str.String()})), nil
}

type parsedVersion struct {
	raw string
}

func (v parsedVersion) LessThan(other values.Value) (bool, *values.Error) {
	otherStr, ok := other.ToString()
	if !ok {
		return false, values.FmtTypeError("parse_version(...).less_than", values.KindString)
	}

	return compareVersions(v.raw, otherStr.String()) < 0, nil
}

func (v parsedVersion) GreaterThan(other values.Value) (bool, *values.Error) {
	otherStr, ok := other.ToString()
	if !ok {
		return false, values.FmtTypeError("parse_version(...).greater_than", values.KindString)
	}

	return compareVersions(v.raw, otherStr.String()) > 0, nil
}

func (v parsedVersion) Matches(spec values.Value) (bool, *values.Error) {
	specStr, ok := spec.ToString()
	if !ok {
		return false, values.FmtTypeError("parse_version(...).matches", values.KindString)
	}

	target := specStr.String()
	if v.raw == target {
		return true, nil
	}

	if hasLetters(v.raw) {
		return false, nil
	}

	return strings.HasPrefix(v.raw, target+"."), nil
}

func compareVersions(a, b string) int {
	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := max(len(partsA), len(partsB))

	for i := range maxLen {
		partA := ""
		partB := ""
		if i < len(partsA) {
			partA = partsA[i]
		}
		if i < len(partsB) {
			partB = partsB[i]
		}

		if cmp := compareVersionPart(partA, partB); cmp != 0 {
			return cmp
		}
	}

	return 0
}

func compareVersionPart(a, b string) int {
	numA, suffixA := parseVersionPart(a)
	numB, suffixB := parseVersionPart(b)

	if numA != numB {
		return numA - numB
	}

	if suffixA == "" && suffixB != "" {
		return 1
	}
	if suffixA != "" && suffixB == "" {
		return -1
	}

	return strings.Compare(suffixA, suffixB)
}

func parseVersionPart(part string) (num int, suffix string) {
	if part == "" {
		return 0, ""
	}

	i := 0
	for i < len(part) && part[i] >= '0' && part[i] <= '9' {
		num = num*10 + int(part[i]-'0')
		i++
	}

	return num, part[i:]
}

func hasLetters(str string) bool {
	for _, c := range str {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			return true
		}
	}
	return false
}
