package common

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common/file"
)

func TestSendHTTPRequest(t *testing.T) {
	// t.Parallel() not used to maintain code coverage for assigning the default
	// HTTPClient.
	methodPost := "pOst"
	methodGet := "GeT"
	methodDelete := "dEleTe"
	methodGarbage := "ding"

	headers := make(map[string]string)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	_, err := SendHTTPRequest(context.Background(),
		methodGarbage, "https://www.google.com", headers,
		strings.NewReader(""), true,
	)
	if err == nil {
		t.Error("Expected error 'invalid HTTP method specified'")
	}
	_, err = SendHTTPRequest(context.Background(),
		methodPost, "https://www.google.com", headers,
		strings.NewReader(""), true,
	)
	if err != nil {
		t.Error(err)
	}
	_, err = SendHTTPRequest(context.Background(),
		methodGet, "https://www.google.com", headers,
		strings.NewReader(""), true,
	)
	if err != nil {
		t.Error(err)
	}

	err = SetHTTPUserAgent("GCTbot/1337.69 (+http://www.lol.com/)")
	if !errors.Is(err, nil) {
		t.Fatalf("received: %v but expected: %v", err, nil)
	}

	_, err = SendHTTPRequest(context.Background(),
		methodDelete, "https://www.google.com", headers,
		strings.NewReader(""), true,
	)
	if err != nil {
		t.Error(err)
	}
	_, err = SendHTTPRequest(context.Background(),
		methodGet, ":missingprotocolscheme", headers,
		strings.NewReader(""), true,
	)
	if err == nil {
		t.Error("Common HTTPRequest accepted missing protocol")
	}
	_, err = SendHTTPRequest(context.Background(),
		methodGet, "test://unsupportedprotocolscheme", headers,
		strings.NewReader(""), true,
	)
	if err == nil {
		t.Error("Common HTTPRequest accepted invalid protocol")
	}
}

func TestSetHTTPClientWithTimeout(t *testing.T) {
	t.Parallel()
	err := SetHTTPClientWithTimeout(-0)
	if !errors.Is(err, errCannotSetInvalidTimeout) {
		t.Fatalf("received: %v but expected: %v", err, errCannotSetInvalidTimeout)
	}

	err = SetHTTPClientWithTimeout(time.Second * 15)
	if !errors.Is(err, nil) {
		t.Fatalf("received: %v but expected: %v", err, nil)
	}
}

func TestSetHTTPUserAgent(t *testing.T) {
	t.Parallel()
	err := SetHTTPUserAgent("")
	if !errors.Is(err, errUserAgentInvalid) {
		t.Fatalf("received: %v but expected: %v", err, errUserAgentInvalid)
	}

	err = SetHTTPUserAgent("testy test")
	if !errors.Is(err, nil) {
		t.Fatalf("received: %v but expected: %v", err, nil)
	}
}

func TestSetHTTPClient(t *testing.T) {
	t.Parallel()
	err := SetHTTPClient(nil)
	if !errors.Is(err, errHTTPClientInvalid) {
		t.Fatalf("received: %v but expected: %v", err, errHTTPClientInvalid)
	}

	err = SetHTTPClient(new(http.Client))
	if !errors.Is(err, nil) {
		t.Fatalf("received: %v but expected: %v", err, nil)
	}
}

func TestIsEnabled(t *testing.T) {
	t.Parallel()
	expected := "Enabled"
	actual := IsEnabled(true)
	if actual != expected {
		t.Errorf("Expected %s. Actual %s", expected, actual)
	}

	expected = "Disabled"
	actual = IsEnabled(false)
	if actual != expected {
		t.Errorf("Expected %s. Actual %s", expected, actual)
	}
}

func TestIsValidCryptoAddress(t *testing.T) {
	t.Parallel()
	b, err := IsValidCryptoAddress("1Mz7153HMuxXTuR2R1t78mGSdzaAtNbBWX", "bTC")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if !b {
		t.Errorf("expected address '%s' to be valid", "1Mz7153HMuxXTuR2R1t78mGSdzaAtNbBWX")
	}

	b, err = IsValidCryptoAddress("bc1qw508d6qejxtdg4y5r3zarvaly0c5xw7kv8f3t4", "bTC")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if !b {
		t.Errorf("expected address '%s' to be valid", "bc1qw508d6qejxtdg4y5r3zarvaly0c5xw7kv8f3t4")
	}

	b, err = IsValidCryptoAddress("an84characterslonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1569pvx", "bTC")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if b {
		t.Errorf("expected address '%s' to be invalid", "an84characterslonghumanreadablepartthatcontainsthenumber1andtheexcludedcharactersbio1569pvx")
	}

	b, err = IsValidCryptoAddress("bc1qc7slrfxkknqcq2jevvvkdgvrt8080852dfjewde450xdlk4ugp7szw5tk9", "bTC")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if !b {
		t.Errorf("expected address '%s' to be valid", "bc1qc7slrfxkknqcq2jevvvkdgvrt8080852dfjewde450xdlk4ugp7szw5tk9")
	}

	b, err = IsValidCryptoAddress("0Mz7153HMuxXTuR2R1t78mGSdzaAtNbBWX", "btc")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if b {
		t.Errorf("expected address '%s' to be invalid", "0Mz7153HMuxXTuR2R1t78mGSdzaAtNbBWX")
	}

	b, err = IsValidCryptoAddress("1Mz7153HMuxXTuR2R1t78mGSdzaAtNbBWX", "lTc")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if b {
		t.Errorf("expected address '%s' to be invalid", "1Mz7153HMuxXTuR2R1t78mGSdzaAtNbBWX")
	}

	b, err = IsValidCryptoAddress("3CDJNfdWX8m2NwuGUV3nhXHXEeLygMXoAj", "ltc")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if !b {
		t.Errorf("expected address '%s' to be valid", "3CDJNfdWX8m2NwuGUV3nhXHXEeLygMXoAj")
	}

	b, err = IsValidCryptoAddress("NCDJNfdWX8m2NwuGUV3nhXHXEeLygMXoAj", "lTc")
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if b {
		t.Errorf("expected address '%s' to be invalid", "NCDJNfdWX8m2NwuGUV3nhXHXEeLygMXoAj")
	}

	b, err = IsValidCryptoAddress(
		"0xb794f5ea0ba39494ce839613fffba74279579268",
		"eth",
	)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if !b {
		t.Errorf("expected address '%s' to be valid", "0xb794f5ea0ba39494ce839613fffba74279579268")
	}

	b, err = IsValidCryptoAddress(
		"xxb794f5ea0ba39494ce839613fffba74279579268",
		"eTh",
	)
	if !errors.Is(err, nil) {
		t.Errorf("received '%v' expected '%v'", err, nil)
	}
	if b {
		t.Errorf("expected address '%s' to be invalid", "xxb794f5ea0ba39494ce839613fffba74279579268")
	}

	b, err = IsValidCryptoAddress(
		"xxb794f5ea0ba39494ce839613fffba74279579268",
		"ding",
	)
	if !errors.Is(err, errInvalidCryptoCurrency) {
		t.Errorf("received '%v' expected '%v'", err, errInvalidCryptoCurrency)
	}
	if b {
		t.Errorf("expected address '%s' to be invalid", "xxb794f5ea0ba39494ce839613fffba74279579268")
	}
}

func TestStringSliceDifference(t *testing.T) {
	t.Parallel()
	originalInputOne := []string{"hello"}
	originalInputTwo := []string{"hello", "moto"}
	expectedOutput := []string{"hello moto"}
	actualResult := StringSliceDifference(originalInputOne, originalInputTwo)
	if reflect.DeepEqual(expectedOutput, actualResult) {
		t.Errorf("Expected '%s'. Actual '%s'",
			expectedOutput, actualResult)
	}
}

func TestStringDataContains(t *testing.T) {
	t.Parallel()
	originalHaystack := []string{"hello", "world", "USDT", "Contains", "string"}
	originalNeedle := "USD"
	anotherNeedle := "thing"
	actualResult := StringDataContains(originalHaystack, originalNeedle)
	if expectedOutput := true; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}
	actualResult = StringDataContains(originalHaystack, anotherNeedle)
	if expectedOutput := false; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}
}

func TestStringDataCompare(t *testing.T) {
	t.Parallel()
	originalHaystack := []string{"hello", "WoRld", "USDT", "Contains", "string"}
	originalNeedle := "WoRld"
	anotherNeedle := "USD"
	actualResult := StringDataCompare(originalHaystack, originalNeedle)
	if expectedOutput := true; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}
	actualResult = StringDataCompare(originalHaystack, anotherNeedle)
	if expectedOutput := false; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}
}

func TestStringDataCompareUpper(t *testing.T) {
	t.Parallel()
	originalHaystack := []string{"hello", "WoRld", "USDT", "Contains", "string"}
	originalNeedle := "WoRld"
	anotherNeedle := "WoRldD"
	actualResult := StringDataCompareInsensitive(originalHaystack, originalNeedle)
	if expectedOutput := true; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}

	actualResult = StringDataCompareInsensitive(originalHaystack, anotherNeedle)
	if expectedOutput := false; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}
}

func TestStringDataContainsUpper(t *testing.T) {
	t.Parallel()
	originalHaystack := []string{"bLa", "BrO", "sUp"}
	originalNeedle := "Bla"
	anotherNeedle := "ning"
	actualResult := StringDataContainsInsensitive(originalHaystack, originalNeedle)
	if expectedOutput := true; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}
	actualResult = StringDataContainsInsensitive(originalHaystack, anotherNeedle)
	if expectedOutput := false; actualResult != expectedOutput {
		t.Errorf("Expected '%v'. Actual '%v'",
			expectedOutput, actualResult)
	}
}

func TestYesOrNo(t *testing.T) {
	t.Parallel()
	if !YesOrNo("y") {
		t.Error("Common YesOrNo Error.")
	}
	if !YesOrNo("yes") {
		t.Error("Common YesOrNo Error.")
	}
	if YesOrNo("ding") {
		t.Error("Common YesOrNo Error.")
	}
}

func TestEncodeURLValues(t *testing.T) {
	t.Parallel()
	urlstring := "https://www.test.com"
	expectedOutput := `https://www.test.com?env=TEST%2FDATABASE&format=json`
	values := url.Values{}
	values.Set("format", "json")
	values.Set("env", "TEST/DATABASE")

	output := EncodeURLValues(urlstring, values)
	if output != expectedOutput {
		t.Error("common EncodeURLValues error")
	}
}

func TestExtractHost(t *testing.T) {
	t.Parallel()
	address := "localhost:1337"
	addresstwo := ":1337"
	expectedOutput := "localhost"
	actualResult := ExtractHost(address)
	if expectedOutput != actualResult {
		t.Errorf(
			"Expected '%s'. Actual '%s'.", expectedOutput, actualResult)
	}
	actualResultTwo := ExtractHost(addresstwo)
	if expectedOutput != actualResultTwo {
		t.Errorf(
			"Expected '%s'. Actual '%s'.", expectedOutput, actualResult)
	}

	address = "192.168.1.100:1337"
	expectedOutput = "192.168.1.100"
	actualResult = ExtractHost(address)
	if expectedOutput != actualResult {
		t.Errorf(
			"Expected '%s'. Actual '%s'.", expectedOutput, actualResult)
	}
}

func TestExtractPort(t *testing.T) {
	t.Parallel()
	address := "localhost:1337"
	expectedOutput := 1337
	actualResult := ExtractPort(address)
	if expectedOutput != actualResult {
		t.Errorf(
			"Expected '%d'. Actual '%d'.", expectedOutput, actualResult)
	}

	address = "localhost"
	expectedOutput = 80
	actualResult = ExtractPort(address)
	if expectedOutput != actualResult {
		t.Errorf(
			"Expected '%d'. Actual '%d'.", expectedOutput, actualResult)
	}
}

func TestGetURIPath(t *testing.T) {
	t.Parallel()
	// mapping of input vs expected result
	testTable := map[string]string{
		"https://api.pro.coinbase.com/accounts":         "/accounts",
		"https://api.pro.coinbase.com/accounts?a=1&b=2": "/accounts?a=1&b=2",
		"http://www.google.com/accounts?!@#$%;^^":       "",
	}
	for testInput, expectedOutput := range testTable {
		actualOutput := GetURIPath(testInput)
		if actualOutput != expectedOutput {
			t.Errorf("Expected '%s'. Actual '%s'.",
				expectedOutput, actualOutput)
		}
	}
}

func TestGetExecutablePath(t *testing.T) {
	t.Parallel()
	if _, err := GetExecutablePath(); err != nil {
		t.Errorf("Common GetExecutablePath. Error: %s", err)
	}
}

func TestGetDefaultDataDir(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		dir, ok := os.LookupEnv("APPDATA")
		if !ok {
			t.Fatal("APPDATA is not set")
		}
		dir = filepath.Join(dir, "GoCryptoTrader")
		actualOutput := GetDefaultDataDir(runtime.GOOS)
		if actualOutput != dir {
			t.Fatalf("Unexpected result. Got: %v Expected: %v", actualOutput, dir)
		}
	default:
		var dir string
		usr, err := user.Current()
		if err == nil {
			dir = usr.HomeDir
		} else {
			var err error
			dir, err = os.UserHomeDir()
			if err != nil {
				dir = "."
			}
		}
		dir = filepath.Join(dir, ".gocryptotrader")
		actualOutput := GetDefaultDataDir(runtime.GOOS)
		if actualOutput != dir {
			t.Fatalf("Unexpected result. Got: %v Expected: %v", actualOutput, dir)
		}
	}
}

func TestCreateDir(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		// test for looking up an invalid directory
		err := CreateDir("")
		if err == nil {
			t.Fatal("expected err due to invalid path, but got nil")
		}

		// test for a directory that exists
		dir, ok := os.LookupEnv("TEMP")
		if !ok {
			t.Fatal("LookupEnv failed. TEMP is not set")
		}
		err = CreateDir(dir)
		if err != nil {
			t.Fatalf("CreateDir failed. Err: %v", err)
		}

		// test for creating a directory
		dir, ok = os.LookupEnv("APPDATA")
		if !ok {
			t.Fatal("LookupEnv failed. APPDATA is not set")
		}
		dir = filepath.Join(dir, "GoCryptoTrader", "TestFileASDFG")
		err = CreateDir(dir)
		if err != nil {
			t.Fatalf("CreateDir failed. Err: %v", err)
		}
		err = os.Remove(dir)
		if err != nil {
			t.Fatalf("Failed to remove file. Err: %v", err)
		}
	default:
		err := CreateDir("")
		if err == nil {
			t.Fatal("expected err due to invalid path, but got nil")
		}

		dir := "/home"
		err = CreateDir(dir)
		if err != nil {
			t.Fatalf("CreateDir failed. Err: %v", err)
		}
		var ok bool
		dir, ok = os.LookupEnv("HOME")
		if !ok {
			t.Fatal("LookupEnv of HOME failed")
		}
		dir = filepath.Join(dir, ".gocryptotrader", "TestFileASFG")
		err = CreateDir(dir)
		if err != nil {
			t.Errorf("CreateDir failed. Err: %s", err)
		}
		err = os.Remove(dir)
		if err != nil {
			t.Fatalf("Failed to remove file. Err: %v", err)
		}
	}
}

func TestChangePermission(t *testing.T) {
	t.Parallel()
	testDir := filepath.Join(os.TempDir(), "TestFileASDFGHJ")
	switch runtime.GOOS {
	case "windows":
		err := ChangePermission("*")
		if err == nil {
			t.Fatal("expected an error on non-existent path")
		}
		err = os.Mkdir(testDir, 0o777)
		if err != nil {
			t.Fatalf("Mkdir failed. Err: %v", err)
		}
		err = ChangePermission(testDir)
		if err != nil {
			t.Fatalf("ChangePerm was unsuccessful. Err: %v", err)
		}
		_, err = os.Stat(testDir)
		if err != nil {
			t.Fatalf("os.Stat failed. Err: %v", err)
		}
		err = os.Remove(testDir)
		if err != nil {
			t.Fatalf("os.Remove failed. Err: %v", err)
		}
	default:
		err := ChangePermission("")
		if err == nil {
			t.Fatal("expected an error on non-existent path")
		}
		err = os.Mkdir(testDir, 0o777)
		if err != nil {
			t.Fatalf("Mkdir failed. Err: %v", err)
		}
		err = ChangePermission(testDir)
		if err != nil {
			t.Fatalf("ChangePerm was unsuccessful. Err: %v", err)
		}
		var a os.FileInfo
		a, err = os.Stat(testDir)
		if err != nil {
			t.Fatalf("os.Stat failed. Err: %v", err)
		}
		if a.Mode().Perm() != file.DefaultPermissionOctal {
			t.Fatalf("expected file permissions differ. expecting file.DefaultPermissionOctal got %#o", a.Mode().Perm())
		}
		err = os.Remove(testDir)
		if err != nil {
			t.Fatalf("os.Remove failed. Err: %v", err)
		}
	}
}

func initStringSlice(size int) (out []string) {
	for x := 0; x < size; x++ {
		out = append(out, "gct-"+strconv.Itoa(x))
	}
	return
}

func TestSplitStringSliceByLimit(t *testing.T) {
	t.Parallel()
	slice50 := initStringSlice(50)
	out := SplitStringSliceByLimit(slice50, 20)
	if len(out) != 3 {
		t.Errorf("expected len() to be 3 instead received: %v", len(out))
	}
	if len(out[0]) != 20 {
		t.Errorf("expected len() to be 20 instead received: %v", len(out[0]))
	}

	out = SplitStringSliceByLimit(slice50, 50)
	if len(out) != 1 {
		t.Errorf("expected len() to be 3 instead received: %v", len(out))
	}
	if len(out[0]) != 50 {
		t.Errorf("expected len() to be 20 instead received: %v", len(out[0]))
	}
}

func TestInArray(t *testing.T) {
	t.Parallel()
	InArray(nil, nil)

	array := [6]int{2, 3, 5, 7, 11, 13}
	isIn, pos := InArray(5, array)
	if !isIn {
		t.Errorf("failed to find the value within the array")
	}
	if pos != 2 {
		t.Errorf("failed return the correct position of the value in the array")
	}
	isIn, _ = InArray(1, array)
	if isIn {
		t.Errorf("found a non existent value in the array")
	}

	slice := make([]int, 0)
	slice = append(append(slice, 5), 3)
	isIn, pos = InArray(5, slice)
	if !isIn {
		t.Errorf("failed to find the value within the slice")
	}
	if pos != 0 {
		t.Errorf("failed return the correct position of the value in the slice")
	}
	isIn, pos = InArray(3, slice)
	if !isIn {
		t.Errorf("failed to find the value within the slice")
	}
	if pos != 1 {
		t.Errorf("failed return the correct position of the value in the slice")
	}
	isIn, _ = InArray(1, slice)
	if isIn {
		t.Errorf("found a non existent value in the slice")
	}
}

func TestErrors(t *testing.T) {
	t.Parallel()
	var test Errors
	if test.Error() != "" {
		t.Fatal("string should be nil")
	}
	errTestOne := errors.New("test1")
	test = append(test, errTestOne)
	if !errors.Is(test, errTestOne) {
		t.Fatal("does not match error")
	}
	test = append(test, errors.New("test2"))
	if test.Error() != "test1, test2" {
		t.Fatal("does not match error")
	}
}

func TestParseStartEndDate(t *testing.T) {
	t.Parallel()
	pt := time.Date(1999, 1, 1, 0, 0, 0, 0, time.Local)
	ft := time.Date(2222, 1, 1, 0, 0, 0, 0, time.Local)
	et := time.Date(2020, 1, 1, 1, 0, 0, 0, time.Local)
	nt := time.Time{}

	err := StartEndTimeCheck(nt, nt)
	if !errors.Is(err, ErrDateUnset) {
		t.Errorf("received %v, expected %v", err, ErrDateUnset)
	}

	err = StartEndTimeCheck(et, nt)
	if !errors.Is(err, ErrDateUnset) {
		t.Errorf("received %v, expected %v", err, ErrDateUnset)
	}

	err = StartEndTimeCheck(et, zeroValueUnix)
	if !errors.Is(err, ErrDateUnset) {
		t.Errorf("received %v, expected %v", err, ErrDateUnset)
	}

	err = StartEndTimeCheck(zeroValueUnix, et)
	if !errors.Is(err, ErrDateUnset) {
		t.Errorf("received %v, expected %v", err, ErrDateUnset)
	}

	err = StartEndTimeCheck(et, zeroValueUnix)
	if !errors.Is(err, ErrDateUnset) {
		t.Errorf("received %v, expected %v", err, ErrDateUnset)
	}

	err = StartEndTimeCheck(zeroValueUnix, et)
	if !errors.Is(err, ErrDateUnset) {
		t.Errorf("received %v, expected %v", err, ErrDateUnset)
	}

	err = StartEndTimeCheck(et, et)
	if !errors.Is(err, ErrStartEqualsEnd) {
		t.Errorf("received %v, expected %v", err, ErrStartEqualsEnd)
	}

	err = StartEndTimeCheck(ft, et)
	if !errors.Is(err, ErrStartAfterTimeNow) {
		t.Errorf("received %v, expected %v", err, ErrStartAfterTimeNow)
	}

	err = StartEndTimeCheck(et, pt)
	if !errors.Is(err, ErrStartAfterEnd) {
		t.Errorf("received %v, expected %v", err, ErrStartAfterEnd)
	}

	err = StartEndTimeCheck(pt, et)
	if !errors.Is(err, nil) {
		t.Errorf("received %v, expected %v", err, nil)
	}
}

func TestGetAssertError(t *testing.T) {
	err := GetAssertError("*[]string", float64(0))
	if err.Error() != "type assert failure from float64 to *[]string" {
		t.Fatal(err)
	}

	err = GetAssertError("<nil>", nil)
	if err.Error() != "type assert failure from <nil> to <nil>" {
		t.Fatal(err)
	}

	err = GetAssertError("bruh", struct{}{})
	if !errors.Is(err, ErrTypeAssertFailure) {
		t.Fatalf("received: '%v' but expected: '%v'", err, ErrTypeAssertFailure)
	}
}

func TestGetAssertError(t *testing.T) {
	err := GetAssertError("*[]string", float64(0))
	if err.Error() != "type assert failure from float64 to *[]string" {
		t.Fatal(err)
	}

	err = GetAssertError("<nil>", nil)
	if err.Error() != "type assert failure from <nil> to <nil>" {
		t.Fatal(err)
	}

	err = GetAssertError("bruh", struct{}{})
	if !errors.Is(err, ErrTypeAssertFailure) {
		t.Fatalf("received: '%v' but expected: '%v'", err, ErrTypeAssertFailure)
	}
}
func TestMatchesEmailPattern(t *testing.T) {
	success := MatchesEmailPattern("someone semail")
	if success {
		t.Error("MatchesEmailPattern() unexpected test validation result")
	}
	success = MatchesEmailPattern("someon esemail@gmail")
	if success {
		t.Error("MatchesEmailPattern() unexpected test validation result")
	}
	success = MatchesEmailPattern("123@gmail")
	if !success {
		t.Error("MatchesEmailPattern() unexpected test validation result")
	}
	success = MatchesEmailPattern("someonesemail@email.com")
	if !success {
		t.Error("MatchesEmailPattern() unexpected test validation result")
	}
}
