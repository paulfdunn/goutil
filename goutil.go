package goutil

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
)

const (
	// MaxUint is the maximum for an uint
	MaxUint = ^uint(0)
	// MinUint is the minimum for an uint
	MinUint = 0

	// MaxInt is the maximum for an int
	MaxInt = int(^uint(0) >> 1)
	// MinInt is the minimum for an int
	MinInt = -MaxInt - 1
)

// ByteSliceToIntSlice converts an byte slice to integer slice
func ByteSliceToIntSlice(bytes []byte) []int {
	out := make([]int, len(bytes))
	for i := range bytes {
		out[i] = int(bytes[i])
	}
	return out
}

// ByteSliceToString converts a byte slice to a hex string with bytesPerLine; no "0x" prefix.
func ByteSliceToString(in []byte, bytesPerLine int) (out string) {
	index := 0
	// Convert bytes to ints, needed for formatting later.
	inInts := make([]int, len(in))
	for i, v := range in {
		inInts[i] = int(v)
	}

	for {
		if index >= len(inInts) {
			break
		}

		// Print bytesPerLine, or a partial line if there are not enough bytes left.
		end := index + bytesPerLine
		if end > len(inInts) {
			end = len(inInts)
		}

		// Ints format nicely with this; space separated.
		s := fmt.Sprintf("%02x", inInts[index:end])
		out += fmt.Sprintf("%s", s[1:len(s)-1])
		out += "\n"
		index += bytesPerLine
	}
	return out
}

// ConvertCamelToUnderscore converts the input string in CamelCase to underscore format.
func ConvertCamelToUnderscore(input string, allLower bool) (output string) {
	for i := range input {
		if len(input) >= i+2 && string(input[i]) == strings.ToLower(string(input[i])) &&
			string(input[i+1]) == strings.ToUpper(string(input[i+1])) {
			// Insert underscore between a lower and upper case character.
			output += string(input[i]) + "_"
		} else {
			output += string(input[i])
		}
	}

	if allLower {
		output = strings.ToLower(output)
	}
	return output
}

// ConvertJSONUnderscoreToCamel converts the input JSON string in underscore format
// to CamelCase; only JSON keys are converted.
func ConvertJSONUnderscoreToCamel(input string) (output string, err error) {
	var inputObject map[string]interface{}
	err = json.Unmarshal([]byte(input), &inputObject)
	if err != nil {
		return "", err
	}

	outputObject, err := ConvertMapUnderscoreToCamel(inputObject)
	if err != nil {
		return "", err
	}
	out, err := json.Marshal(outputObject)
	return string(out), err
}

// ConvertMapUnderscoreToCamel converts the input JSON map in underscore format
// to CamelCase; only JSON keys are converted.
func ConvertMapUnderscoreToCamel(input map[string]interface{}) (output map[string]interface{}, err error) {
	output = make(map[string]interface{})
	for k, v := range input {
		if newV, ok := v.(map[string]interface{}); ok {
			output[ConvertUnderscoreToCamel(k)], err = ConvertMapUnderscoreToCamel(newV)
			if err != nil {
				return output, err
			}
		} else if newV, ok := v.([]interface{}); ok {
			out := make([]interface{}, 0)
			for _, nv := range newV {
				s, err := json.Marshal(nv)
				if err != nil {
					return output, err
				}

				o, _ := ConvertJSONUnderscoreToCamel(string(s))
				var sObj interface{}
				err = json.Unmarshal([]byte(o), &sObj)
				if err != nil {
					return output, err
				}
				out = append(out, sObj)
			}
			output[ConvertUnderscoreToCamel(k)] = out
		} else {
			output[ConvertUnderscoreToCamel(k)] = v
		}
	}

	return output, nil
}

// ConvertUnderscoreToCamel converts a single input word from underscore format to CamelCase.
func ConvertUnderscoreToCamel(input string) (output string) {
	for i := range input {
		if i == 0 && string(input[i]) != "_" {
			// Capitalize first character if not underscore.
			output += strings.ToUpper(string(input[i]))
		} else if i == 0 && string(input[i]) == "_" {
			// Skip leading underscore.
		} else if i == 1 && string(input[i]) != "_" && string(input[i-1]) == "_" {
			// Capitalize character after a leading underscore.
			output += strings.ToUpper(string(input[i]))
		} else if i >= 2 && string(input[i]) != "_" &&
			string(input[i-1]) == "_" && string(input[i-2]) != "_" {
			// Capitalize character after a underscore, where underscore is precedeed by
			// non-underscore.
			output += strings.ToUpper(string(input[i]))
		} else if string(input[i]) == "_" {
			// Skip underscores in output.
		} else {
			output += string(input[i])
		}
	}

	// Abbreviations will be all caps.
	abbreviations := []string{"JSON", "NQN", "HTTP"}
	for _, abrv := range abbreviations {
		output = regexp.MustCompile(fmt.Sprintf(`(?i)(%s)`, abrv)).ReplaceAllString(output, abrv)
	}

	return output
}

// DirIsEmpty returns true if the directory exists and is empty.
func DirIsEmpty(path string) (bool, error) {
	f, err := os.Open(path)
	defer f.Close()
	if err != nil {
		// Return false if the directory does not exist.
		return false, err
	}

	// Readdirnames does NOT return "." and ".."; so a single file indicates the dir
	// is not empty.
	_, err = f.Readdirnames(1)
	if err == io.EOF {
		return true, nil
	}

	return false, err
}

// EnumsFromMapIntString creates lists of keys and values from a map[int]string.
func EnumsFromMapIntString(m map[int]string) (keys []int, values []string) {
	keys = make([]int, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Ints(keys)

	for _, k := range keys {
		values = append(values, m[k])
	}

	return keys, values
}

// InIntSlice checks if a int slice contains specific int.
func InIntSlice(intToFind int, list []int) bool {
	for _, v := range list {
		if v == intToFind {
			return true
		}
	}
	return false
}

// InStringSlice checks if a string slice contains specific string.
func InStringSlice(stringToFind string, list []string) bool {
	for _, v := range list {
		if v == stringToFind {
			return true
		}
	}
	return false
}

// InStringSlicePtr checks if a string slice contains specific string.
func InStringSlicePtr(stringToFind string, list []*string) bool {
	values := make([]string, 0)
	for _, v := range list {
		values = append(values, *v)
	}
	return InStringSlice(stringToFind, values)
}

// IntSliceIsASCII tests an integer slice to see if all values are in the printable
// ASCII range; returns true if yes, false otherwise.
// filter is used to filter out values, like 0 that is
// used for padding some string and is used for EOF; also used
// to filter CR/LF/etc.
// filter is a map mainly so text can be added to describe why a value is
// filtered out, but the text is not required.
// If the input contains ONLY filtered values, returns false
// Erors if the input is an empty slice, or if all input values are filtered out.
func IntSliceIsASCII(in []int, filter map[int]string) (bool, error) {
	min, max, err := MinMaxIntSlice(in, filter)
	// Cannot include 127, which is DEL, or text compares on binary
	// data will fail when a single field includes 0x7f
	if min >= 32 && max < 127 {
		return true, err
	}
	return false, err
}

// IntSliceRemoveDuplicates removes duplicates from to integer slices; results may not be stable.
func IntSliceRemoveDuplicates(in []int) []int {
	// Merge, and use a map to eliminate duplicates.
	m := map[int]int{}
	for i := range in {
		if _, ok := m[in[i]]; ok {
			m[in[i]]++
		} else {
			m[in[i]] = 1
		}
	}
	var index int
	newAll := make([]int, len(m))
	for k := range m {
		newAll[index] = k
		index++
	}
	return newAll
}

// MD5Checksum provides a []byte with the MD5 hash (checksum) for the input.
func MD5Checksum(input []byte) [16]byte {
	return md5.Sum(input)
}

// MD5ChecksumBase64 provides a string with the MD5 hash (checksum) in base64 for the input.
func MD5ChecksumBase64(input []byte) string {
	s := MD5Checksum(input)
	return base64.StdEncoding.EncodeToString(s[:])
}

// MinMaxIntSlice returns the max and min for an int slice.
// Errors if the input is an empty slice, or if all input values are filtered out.
// filter is used to filter out specific values; it is a map mainly
// so text can be added to describe why a value is filtered out,
// but the text is not required.
func MinMaxIntSlice(in []int, filter map[int]string) (int, int, error) {
	var max = MinInt
	var min = MaxInt
	var found = false
	var err error
	for _, value := range in {
		// Skip values in the filter
		if _, ok := filter[value]; ok {
			continue
		}
		found = true
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}

	if !found {
		err = errors.New("MinMaxIntSlice: all inputs were filtered")
	}

	return min, max, err
}

// PrettyJSON transforms JSON for more friendly screen output.
// Transforms this:
// "SomeJSONField": [1,
//
//	2,
//	3,
//	4      ],
//
// into:
// "SomeJSONField": [1,2,3,4],
func PrettyJSON(json []byte) []byte {
	// re1: remove all CRLF from lines that only have a number followed by
	// comma. This gets rid of all CRLF, but leaves the initial CRLF
	// after the opening "["
	re1 := regexp.MustCompile(`(?m:^\s*?([0-9.]+,?)\s*?\r?\n?)`)
	json = re1.ReplaceAll(json, []byte("$1"))
	// re:2 Now get rid of the CRLF immediately after "[" if it is followed by a
	//  number and comma.
	re2 := regexp.MustCompile(`(?m:\[\s*?\r?\n?([0-9.]+,)\r?\n?)`)
	json = re2.ReplaceAll(json, []byte("[$1"))
	// re3: remove the trailing spaces after the final number and prior to the final "]"
	re3 := regexp.MustCompile(`([0-9.])\s*?]`)
	json = re3.ReplaceAll(json, []byte("$1]"))

	// JSON converts actual \n to "\n"; undo that
	re4 := regexp.MustCompile(`\n`)
	json = re4.ReplaceAll(json, []byte("\n"))

	// Remove trailing whitespace from any line so that output is
	// compatible with Golang Examples.
	re5 := regexp.MustCompile(`(?m)\s*?$`)
	json = re5.ReplaceAll(json, []byte(""))

	return json
}

// RequestUsername will return the username of the request when using basic or digest
// authentication; if it can be determined.
func RequestUsername(r *http.Request) string {
	// r.Header["Authorization"] is a slice of strings. I.E.
	// Basic authentication.
	// "Authorization":[]string{"Basic YWRtaW46YWRtaW4="},
	// Digest authentication
	// r.Header["Authorization"] is a slice of strings. I.E.
	// "Authorization":[]string{"Digest username=\"admin\", realm=\"Western Digital Corporation\", nonce=\"AHYBbBIPrPRMzsDo\",...}
	for _, v := range r.Header["Authorization"] {
		splits := strings.Split(v, ",")
		for _, split := range splits {
			if strings.Contains(split, "username") {
				u := strings.Split(split, "=")
				if len(u) == 2 {
					return strings.Replace(u[1], `"`, ``, -1)
				}

				return ""
			} else if strings.Contains(split, "Basic ") {
				u := strings.Split(split, " ")
				if len(u) == 2 {
					user, _ := base64.StdEncoding.DecodeString(u[1])
					userSplit := strings.Split(string(user), ":")
					return string(userSplit[0])
				}

				return ""
			}
		}
	}

	return ""
}

// Round a number to the nearest number of digits; I.E. 0 to round
// to an integer.
func Round(x float64, digits int) float64 {
	return math.Floor(x*math.Pow10(digits)+0.5) / math.Pow10(digits)
}

// SHA1Checksum provides a []byte with the MD5 hash (checksum) for the input.
func SHA1Checksum(input []byte) [20]byte {
	return sha1.Sum(input)
}

// SHA1ChecksumBase64 provides a string with the MD5 hash (checksum) in base64 for the input.
func SHA1ChecksumBase64(input []byte) string {
	s := SHA1Checksum(input)
	return base64.StdEncoding.EncodeToString(s[:])
}

// UniqueStrings creates a list of unique strings from the input.
// Pass in a slice of  strings. Each string is checked against the value
// of prior strings in the list, and a "_#" appended if required to make the name unique.
// A new list is returned, as well as a boolean indicating if any duplicates
// occurred.
// If any inputs reduce to "" with strings.TrimSpace, they are replaced with "_"
// and then above appends are added to create unique names. This prevents a
// entirely blank string being returned.
func UniqueStrings(input []string, numberFormat string) ([]string, bool) {
	m := map[string]int{}
	duplicates := false
	ret := make([]string, len(input))
	for i := range input {
		// Dont allow an empty string to be returned.
		if strings.TrimSpace(input[i]) == "" {
			input[i] = "_"
		}
		if _, ok := m[input[i]]; ok {
			m[input[i]]++
			ret[i] = fmt.Sprintf(numberFormat, input[i], m[input[i]])
			duplicates = true
		} else {
			m[input[i]] = 1
			ret[i] = input[i]
		}
	}

	return ret, duplicates
}

// VerifyMapKeysStringString verifies an input map contains required keys;
// true is all keys found, false otherwise.
func VerifyMapKeysStringString(keys []string, testMap map[string]string) bool {
	allKeysFound := true
	for i := range keys {
		if _, ok := testMap[keys[i]]; !ok {
			allKeysFound = false
			break
		}
	}
	return allKeysFound
}
