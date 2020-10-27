package goutil

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"sort"
	"strings"
	"testing"
)

func init() {
}

var (
	reqUser string
)

func ExampleByteSliceToString() {
	fmt.Printf(ByteSliceToString([]byte{}, 3))
	fmt.Printf(ByteSliceToString([]byte{0}, 3))
	fmt.Printf(ByteSliceToString([]byte{0, 1, 2}, 3))
	fmt.Printf(ByteSliceToString([]byte{0, 1, 2, 3}, 3))
	fmt.Printf(ByteSliceToString([]byte{0, 1, 2, 3, 4, 5}, 3))

	// Output:
	// 00
	// 00 01 02
	// 00 01 02
	// 03
	// 00 01 02
	// 03 04 05
}

func ExampleConvertCamelToUnderscore() {
	fmt.Println(ConvertCamelToUnderscore("CamelCase", false))
	fmt.Println(ConvertCamelToUnderscore("CamelCase", true))
	fmt.Println(ConvertCamelToUnderscore("MULtipleLeading", true))
	fmt.Println(ConvertCamelToUnderscore("SingleEndC", true))

	// Output:
	// Camel_Case
	// camel_case
	// multiple_leading
	// single_end_c
}

func ExampleConvertJSONUnderscoreToCamel() {
	// VOLUMES_EXIST_ON_SET is part of the message; it will not be changed.
	i, _ := ConvertJSONUnderscoreToCamel(`{"jsonrpc":"2.0","id":1,"error":{"code":10,"message":"VOLUMES_EXIST_ON_SET","want_camel":1}}`)
	fmt.Printf("%+v\n", i)

	i, _ = ConvertJSONUnderscoreToCamel(`{"Sets":[{"SetID":0,"TotalBytes":85899345920,"FreeBytes":0}]}`)
	fmt.Printf("%+v\n", i)

	// Output:
	// {"Error":{"Code":10,"Message":"VOLUMES_EXIST_ON_SET","WantCamel":1},"Id":1,"JSONrpc":"2.0"}
	// {"Sets":[{"FreeBytes":0,"SetID":0,"TotalBytes":85899345920}]}
}

func ExampleConvertMapUnderscoreToCamel() {
	m, _ := ConvertMapUnderscoreToCamel(map[string]interface{}{"some_key": 1})
	fmt.Printf("Result:%+v\n", m)

	inputRecurssive := map[string]interface{}{"some_key": 1, "a_map": map[string]interface{}{"some_key1": 1}}
	m, _ = ConvertMapUnderscoreToCamel(inputRecurssive)
	// Print by sorted keys so the output is always in the same order.
	keys := make([]string, 0)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%+v:%+v\n", k, m[k])
	}

	// Output:
	// Result:map[SomeKey:1]
	// AMap:map[SomeKey1:1]
	// SomeKey:1
}

func ExampleConvertUnderscoreToCamel() {
	fmt.Println(ConvertUnderscoreToCamel("_leading_underscore"))
	fmt.Println(ConvertUnderscoreToCamel("a_leading_underscore"))
	fmt.Println(ConvertUnderscoreToCamel("camel_case"))

	// Output:
	// LeadingUnderscore
	// ALeadingUnderscore
	// CamelCase
}

func ExampleDirIsEmpty() {
	u, _ := user.Current()
	b, _ := DirIsEmpty(u.HomeDir)
	fmt.Printf("User dir is empty? %+v\n", b)

	tmpDir, _ := ioutil.TempDir("", "")
	b, _ = DirIsEmpty(tmpDir)
	fmt.Printf("Temp dir is empty? %+v\n", b)
	os.Remove(tmpDir)
	// Output:
	// User dir is empty? false
	// Temp dir is empty? true
}

func ExampleEnumsFromMapIntString() {
	m := map[int]string{1: "one", 2: "two"}
	k, v := EnumsFromMapIntString(m)
	fmt.Printf("%+v %+v", k, v)
	// Output:
	// [1 2] [one two]
}

// Test without the use of a filter
func ExampleIntSliceIsASCII_true() {
	someInts := []int{32, 33, 125, 126}
	filter := map[int]string{}
	ascii, _ := IntSliceIsASCII(someInts, filter)
	fmt.Printf("IntSliceIsASCII:%v", ascii)
	// Output:
	// IntSliceIsASCII:true
}

func ExampleInIntSlice() {
	fmt.Println(InIntSlice(0, []int{1, 2, 3, 4, 5}))
	fmt.Println(InIntSlice(3, []int{1, 2, 3, 4, 5}))
	// Output:
	// false
	// true
}

func ExampleInStringSlice() {
	fmt.Println(InStringSlice("hello", []string{"nothello", "goodbye"}))
	fmt.Println(InStringSlice("hello", []string{"hello", "goodbye"}))
	// Output:
	// false
	// true
}

func ExampleInStringSlicePtr() {
	h := "hello"
	nh := "nothello"
	g := "goodbye"
	fmt.Println(InStringSlicePtr("hello", []*string{&nh, &g}))
	fmt.Println(InStringSlicePtr("hello", []*string{&h, &g}))
	// Output:
	// false
	// true
}

// Test with the use of a filter
func ExampleIntSliceIsASCII_false() {
	someInts := []int{10, 1, -50, 1000, -10, -1, 50, -1000}
	filter := map[int]string{}
	ascii, _ := IntSliceIsASCII(someInts, filter)
	fmt.Printf("IntSliceIsASCII:%v", ascii)
	// Output:
	// IntSliceIsASCII:false
}

// Test the case where all input values are also included in the filter; should panic.
func ExampleIntSliceIsASCII_allValuesFiltered() {
	someInts := []int{32, 33, 126, 127}
	filter := map[int]string{32: "", 33: "", 126: "", 127: ""}
	_, err := IntSliceIsASCII(someInts, filter)
	fmt.Printf("Error:%v", err)
	// Output:
	// Error:MinMaxIntSlice: all inputs were filtered
}

func ExampleIntSliceRemoveDuplicates() {
	r := IntSliceRemoveDuplicates([]int{1, 2, 3, 4, 4, 1, 7, 8})
	// Result may not be stable, so sort prior to Output.
	sort.Sort(sort.IntSlice(r))
	fmt.Println(r)
	// Output:
	// [1 2 3 4 7 8]
}

func ExampleMD5ChecksumBase64() {
	fmt.Print(fmt.Sprintf("%s", MD5ChecksumBase64([]byte("admin:Western Digital Corporation:admin"))))
	// Output:
	// l+uthS0Nq/1rca4m//Yfow==
}

func ExampleMD5Checksum() {
	fmt.Print(fmt.Sprintf("% 02x", MD5Checksum([]byte("admin:Western Digital Corporation:admin"))))
	// Output:
	// 97 eb ad 85 2d 0d ab fd 6b 71 ae 26 ff f6 1f a3
}

// Test with the use of a filter
func ExampleMinMaxIntSlice() {
	someInts := []int{10, 1, -50, 1000, -10, -1, 50, -1000}
	filter := map[int]string{-1000: "don't like this value", 1000: "or this"}
	min, max, _ := MinMaxIntSlice(someInts, filter)
	fmt.Printf("min:%v, max:%v", min, max)
	// Output:
	// min:-50, max:50
}

// Test without the use of a filter.
func ExampleMinMaxIntSlice_noFilter() {
	someInts := []int{10, 1, -50, 1000, -10, -1, 50, -1000}
	filter := map[int]string{}
	min, max, _ := MinMaxIntSlice(someInts, filter)
	fmt.Printf("min:%v, max:%v", min, max)
	// Output:
	// min:-1000, max:1000
}

// Test with nil filter
func ExampleMinMaxIntSlice_nilFilter() {
	someInts := []int{10, 1, -50, 1000, -10, -1, 50, -1000}
	min, max, _ := MinMaxIntSlice(someInts, nil)
	fmt.Printf("min:%v, max:%v", min, max)
	// Output:
	// min:-1000, max:1000
}

// Test the case where all input values are also included in the filter; should panic.
func ExampleMinMaxIntSlice_allValuesFiltered() {
	someInts := []int{10, 1, -50, 1000, -10, -1, 50, -1000}
	filter := map[int]string{10: "", 1: "", -50: "", 1000: "", -10: "", -1: "", 50: "", -1000: ""}
	_, _, err := MinMaxIntSlice(someInts, filter)
	fmt.Printf("Error:%v", err)
	// Output:
	// Error:MinMaxIntSlice: all inputs were filtered
}

func ExamplePrettyJSON() {
	testJSON := []byte(
		`"field": [
1,
2,
3    ],`)
	pj := PrettyJSON(testJSON)
	fmt.Println(string(pj))
	// Output:
	// "field": [1,2,3],
}

func ExampleRound_pi0() {
	var rounded float64
	rounded = Round(math.Pi, 0)
	fmt.Printf("%.0f", rounded)
	// Output:
	// 3
}

func ExampleRound_pi1() {
	var rounded float64
	rounded = Round(math.Pi, 1)
	fmt.Printf("%.1f", rounded)
	// Output:
	// 3.1
}

func ExampleRound_pi5() {
	var rounded float64
	rounded = Round(math.Pi, 5)
	fmt.Printf("%.5f", rounded)
	// Output:
	// 3.14159
}

func ExampleRound_n2l() {
	var rounded float64
	rounded = Round(1.494, 2)
	fmt.Printf("%.2f", rounded)
	// Output:
	// 1.49
}

func ExampleRound_n2h() {
	var rounded float64
	rounded = Round(1.495, 2)
	fmt.Printf("%.2f", rounded)
	// Output:
	// 1.50
}

func ExampleSHA1ChecksumBase64() {
	fmt.Print(fmt.Sprintf("%s", SHA1ChecksumBase64([]byte("admin"))))
	// Output:
	// 0DPiKuNIrrVmD8IUCuw1hQxNqZc=
}

func ExampleSHA1Checksum() {
	fmt.Print(fmt.Sprintf("% 02x", SHA1Checksum([]byte("admin"))))
	// Output:
	// d0 33 e2 2a e3 48 ae b5 66 0f c2 14 0a ec 35 85 0c 4d a9 97
}

func ExampleUniqueStrings() {
	s := []string{"paul", "paul", "bruce", "jeff", "bruce", "bruce", "bob", "paul", "", ""}
	o, b := UniqueStrings(s, "%s_%03d")
	fmt.Println(strings.Join(o, "|"), b)

	s = []string{"paul", "bruce", "jeff"}
	o, b = UniqueStrings(s, "%s_%03d")
	fmt.Println(strings.Join(o, "|"), b)

	// Output:
	// paul|paul_002|bruce|jeff|bruce_002|bruce_003|bob|paul_003|_|__002 true
	// paul|bruce|jeff false
}

func ExampleVerifyMapKeysStringString() {
	kf := []string{"1", "4"}
	kt := []string{"1", "2"}
	m := map[string]string{"1": "one", "2": "two"}
	fmt.Printf("Missing key:%v\n", VerifyMapKeysStringString(kf, m))
	fmt.Printf("Contains keys:%v", VerifyMapKeysStringString(kt, m))

	// Output:
	// Missing key:false
	// Contains keys:true
}

func TestRequestUsername(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(testHandlerFuncUser))
	defer ts.Close()

	handler := http.HandlerFunc(testHandlerFuncUser)
	resp := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "http://TestRequestUsername", nil)
	req.SetBasicAuth("testUser", "password")
	handler.ServeHTTP(resp, req)

	if reqUser != "testUser" {
		t.Errorf("User was not correct, reqUser:%+v", reqUser)
	} else {
		//fmt.Printf("User was:%+v", reqUser)
	}
}

func testHandlerFuncUser(w http.ResponseWriter, r *http.Request) {
	reqUser = RequestUsername(r)
}
