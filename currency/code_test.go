package currency

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestRoleString(t *testing.T) {
	if Unset.String() != UnsetRoleString {
		t.Errorf("Role String() error expected %s but received %s",
			UnsetRoleString,
			Unset)
	}

	if Fiat.String() != FiatCurrencyString {
		t.Errorf("Role String() error expected %s but received %s",
			FiatCurrencyString,
			Fiat)
	}

	if Cryptocurrency.String() != CryptocurrencyString {
		t.Errorf("Role String() error expected %s but received %s",
			CryptocurrencyString,
			Cryptocurrency)
	}

	if Token.String() != TokenString {
		t.Errorf("Role String() error expected %s but received %s",
			TokenString,
			Token)
	}

	if Contract.String() != ContractString {
		t.Errorf("Role String() error expected %s but received %s",
			ContractString,
			Contract)
	}

	var random Role = 1 << 7

	if random.String() != UnsetRoleString {
		t.Errorf("Role String() error expected %s but received %s",
			"UNKNOWN",
			random)
	}
}

func TestRoleMarshalJSON(t *testing.T) {
	d, err := json.Marshal(Fiat)
	if err != nil {
		t.Error("Role MarshalJSON() error", err)
	}

	if expected := `"fiatcurrency"`; string(d) != expected {
		t.Errorf("Role MarshalJSON() error expected %s but received %s",
			expected,
			string(d))
	}
}

// TestRoleUnmarshalJSON logic test
func TestRoleUnmarshalJSON(t *testing.T) {
	type AllTheRoles struct {
		RoleOne     Role `json:"RoleOne"`
		RoleTwo     Role `json:"RoleTwo"`
		RoleThree   Role `json:"RoleThree"`
		RoleFour    Role `json:"RoleFour"`
		RoleFive    Role `json:"RoleFive"`
		RoleSix     Role `json:"RoleSix"`
		RoleUnknown Role `json:"RoleUnknown"`
	}

	var outgoing = AllTheRoles{
		RoleOne:   Unset,
		RoleTwo:   Cryptocurrency,
		RoleThree: Fiat,
		RoleFour:  Token,
		RoleFive:  Contract,
		RoleSix:   Stable,
	}

	e, err := json.Marshal(1337)
	if err != nil {
		t.Fatal("Role UnmarshalJSON() error", err)
	}

	var incoming AllTheRoles
	err = json.Unmarshal(e, &incoming)
	if err == nil {
		t.Fatal("Role UnmarshalJSON() Expected error")
	}

	e, err = json.Marshal(outgoing)
	if err != nil {
		t.Fatal("Role UnmarshalJSON() error", err)
	}

	err = json.Unmarshal(e, &incoming)
	if err != nil {
		t.Fatal("Role UnmarshalJSON() error", err)
	}

	if incoming.RoleOne != Unset {
		t.Errorf("Role String() error expected %s but received %s",
			Unset,
			incoming.RoleOne)
	}

	if incoming.RoleTwo != Cryptocurrency {
		t.Errorf("Role String() error expected %s but received %s",
			Cryptocurrency,
			incoming.RoleTwo)
	}

	if incoming.RoleThree != Fiat {
		t.Errorf("Role String() error expected %s but received %s",
			Fiat,
			incoming.RoleThree)
	}

	if incoming.RoleFour != Token {
		t.Errorf("Role String() error expected %s but received %s",
			Token,
			incoming.RoleFour)
	}

	if incoming.RoleFive != Contract {
		t.Errorf("Role String() error expected %s but received %s",
			Contract,
			incoming.RoleFive)
	}

	if incoming.RoleUnknown != Unset {
		t.Errorf("Role String() error expected %s but received %s",
			incoming.RoleFive,
			incoming.RoleUnknown)
	}
	var unhandled Role
	err = unhandled.UnmarshalJSON([]byte("\"ThisIsntReal\""))
	if err == nil {
		t.Error("Expected unmarshall error")
	}

	err = unhandled.UnmarshalJSON([]byte(`1336`))
	if err == nil {
		t.Error("Expected unmarshall error")
	}
}

func TestBaseCode(t *testing.T) {
	var main BaseCodes
	if main.HasData() {
		t.Errorf("BaseCode HasData() error expected false but received %v",
			main.HasData())
	}

	_ = main.Register("CATS", Unset)
	if !main.HasData() {
		t.Errorf("BaseCode HasData() error expected true but received %v",
			main.HasData())
	}

	// Changes unset to fiat
	catsCode := main.Register("CATS", Fiat)
	if !main.HasData() {
		t.Errorf("BaseCode HasData() error expected true but received %v",
			main.HasData())
	}

	_ = main.Register("CATS", Stable)
	if !main.HasData() {
		t.Errorf("BaseCode HasData() error expected true but received %v",
			main.HasData())
	}

	if !main.Register("CATS", Unset).Equal(catsCode) {
		t.Errorf("BaseCode Match() error expected true but received %v",
			false)
	}

	if main.Register("DOGS", Unset).Equal(catsCode) {
		t.Errorf("BaseCode Match() error expected false but received %v",
			true)
	}

	loadedCurrencies := main.GetCurrencies()

	if loadedCurrencies.Contains(main.Register("OWLS", Unset)) {
		t.Errorf("BaseCode Contains() error expected false but received %v",
			true)
	}

	if !loadedCurrencies.Contains(catsCode) {
		t.Errorf("BaseCode Contains() error expected true but received %v",
			false)
	}

	main.Register("XBTUSD", Unset)

	err := main.UpdateCurrency("Bitcoin Perpetual",
		"XBTUSD",
		"",
		0,
		Contract)
	if err != nil {
		t.Fatal(err)
	}

	main.Register("BTC", Unset)
	err = main.UpdateCurrency("Bitcoin", "BTC", "", 1337, Unset)
	if !errors.Is(err, errRoleUnset) {
		t.Fatalf("received: '%v' but expected: '%v'", err, errRoleUnset)
	}

	err = main.UpdateCurrency("Bitcoin", "BTC", "", 1337, Cryptocurrency)
	if err != nil {
		t.Fatal(err)
	}

	aud := main.Register("AUD", Unset)
	err = main.UpdateCurrency("Unreal Dollar", "AUD", "", 1111, Fiat)
	if err != nil {
		t.Fatal(err)
	}

	if aud.Item.FullName != "Unreal Dollar" {
		t.Error("Expected fullname to update for AUD")
	}

	err = main.UpdateCurrency("Australian Dollar", "AUD", "", 1336, Fiat)
	if err != nil {
		t.Fatal(err)
	}

	aud.Item.Role = Unset
	err = main.UpdateCurrency("Australian Dollar", "AUD", "", 1336, Fiat)
	if err != nil {
		t.Fatal(err)
	}
	if aud.Item.Role != Fiat {
		t.Error("Expected role to change to Fiat")
	}

	main.Register("PPT", Unset)
	err = main.UpdateCurrency("Populous", "PPT", "ETH", 1335, Token)
	if err != nil {
		t.Fatal(err)
	}

	contract := main.Register("XBTUSD", Unset)

	if contract.IsFiatCurrency() {
		t.Errorf("BaseCode IsFiatCurrency() error expected false but received %v",
			true)
	}

	if contract.IsCryptocurrency() {
		t.Errorf("BaseCode IsCryptocurrency() error expected false but received %v",
			true)
	}

	err = main.LoadItem(nil)
	if !errors.Is(err, errItemIsNil) {
		t.Fatalf("received: '%v' but expected: '%v'", err, errItemIsNil)
	}

	err = main.LoadItem(&Item{})
	if !errors.Is(err, errItemIsEmpty) {
		t.Fatalf("received: '%v' but expected: '%v'", err, errItemIsEmpty)
	}

	err = main.LoadItem(&Item{
		ID:       0,
		FullName: "Cardano",
		Role:     Cryptocurrency,
		Symbol:   "ADA",
	})
	if err != nil {
		t.Fatal(err)
	}

	err = main.LoadItem(&Item{
		ID:       0,
		FullName: "Cardano",
		Role:     Cryptocurrency,
		Symbol:   "ADA",
	})
	if err != nil {
		t.Fatal(err)
	}

	full, err := main.GetFullCurrencyData()
	if err != nil {
		t.Error("BaseCode GetFullCurrencyData error", err)
	}

	if len(full.Contracts) != 1 {
		t.Errorf("BaseCode GetFullCurrencyData() error expected 1 but received %v",
			len(full.Contracts))
	}

	if len(full.Cryptocurrency) != 2 {
		t.Errorf("BaseCode GetFullCurrencyData() error expected 1 but received %v",
			len(full.Cryptocurrency))
	}

	if len(full.FiatCurrency) != 2 {
		t.Errorf("BaseCode GetFullCurrencyData() error expected 1 but received %v",
			len(full.FiatCurrency))
	}

	if len(full.Token) != 1 {
		t.Errorf("BaseCode GetFullCurrencyData() error expected 1 but received %v",
			len(full.Token))
	}

	if len(full.UnsetCurrency) != 2 {
		t.Errorf("BaseCode GetFullCurrencyData() error expected 3 but received %v",
			len(full.UnsetCurrency))
	}

	if full.LastMainUpdate.(int64) != -62135596800 {
		t.Errorf("BaseCode GetFullCurrencyData() error expected -62135596800 but received %d",
			full.LastMainUpdate)
	}

	err = main.LoadItem(&Item{
		ID:       0,
		FullName: "Cardano",
		Role:     Role(99),
		Symbol:   "ADA",
	})
	if err != nil {
		t.Error("BaseCode LoadItem() error", err)
	}
	_, err = main.GetFullCurrencyData()
	if err == nil {
		t.Error("Expected 'Role undefined'")
	}

	main.Items[0].FullName = "Hello"
	err = main.UpdateCurrency("MEWOW", "CATS", "", 1338, Fiat)
	if err != nil {
		t.Fatal(err)
	}

	if main.Items[0].FullName != "MEWOW" {
		t.Error("Fullname not updated")
	}
	err = main.UpdateCurrency("MEWOW", "CATS", "", 1338, Fiat)
	if err != nil {
		t.Fatal(err)
	}
	err = main.UpdateCurrency("WOWCATS", "CATS", "", 3, Fiat)
	if err != nil {
		t.Fatal(err)
	}

	// Creates a new item under a different currency role
	if main.Items[0].ID != 3 {
		t.Error("ID not updated")
	}

	main.Items[0].Role = Unset
	err = main.UpdateCurrency("MEWOW", "CATS", "", 1338, Cryptocurrency)
	if err != nil {
		t.Fatal(err)
	}
	if main.Items[0].ID != 1338 {
		t.Error("ID not updated")
	}
}

func TestCodeString(t *testing.T) {
	if cc, expected := NewCode("TEST"), "TEST"; cc.String() != expected {
		t.Errorf("Currency Code String() error expected %s but received %s",
			expected, cc)
	}
}

func TestCodeLower(t *testing.T) {
	if cc, expected := NewCode("TEST"), "test"; cc.Lower().String() != expected {
		t.Errorf("Currency Code Lower() error expected %s but received %s",
			expected,
			cc.Lower())
	}
}

func TestCodeUpper(t *testing.T) {
	if cc, expected := NewCode("test"), "TEST"; cc.Upper().String() != expected {
		t.Errorf("Currency Code Upper() error expected %s but received %s",
			expected,
			cc.Upper())
	}
}

func TestCodeUnmarshalJSON(t *testing.T) {
	var unmarshalHere Code
	expected := "BRO"
	encoded, err := json.Marshal(expected)
	if err != nil {
		t.Fatal("Currency Code UnmarshalJSON error", err)
	}

	err = json.Unmarshal(encoded, &unmarshalHere)
	if err != nil {
		t.Fatal("Currency Code UnmarshalJSON error", err)
	}

	err = json.Unmarshal(encoded, &unmarshalHere)
	if err != nil {
		t.Fatal("Currency Code UnmarshalJSON error", err)
	}

	if unmarshalHere.String() != expected {
		t.Errorf("Currency Code Upper() error expected %s but received %s",
			expected,
			unmarshalHere)
	}

	encoded, err = json.Marshal(1336) // :'(
	if err != nil {
		t.Fatal(err)
	}
	err = json.Unmarshal(encoded, &unmarshalHere)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCodeMarshalJSON(t *testing.T) {
	quickstruct := struct {
		Codey Code `json:"sweetCodes"`
	}{
		Codey: NewCode("BRO"),
	}

	expectedJSON := `{"sweetCodes":"BRO"}`

	encoded, err := json.Marshal(quickstruct)
	if err != nil {
		t.Fatal("Currency Code UnmarshalJSON error", err)
	}

	if string(encoded) != expectedJSON {
		t.Errorf("Currency Code Upper() error expected %s but received %s",
			expectedJSON,
			string(encoded))
	}

	quickstruct = struct {
		Codey Code `json:"sweetCodes"`
	}{
		Codey: Code{}, // nil code
	}

	encoded, err = json.Marshal(quickstruct)
	if err != nil {
		t.Fatal("Currency Code UnmarshalJSON error", err)
	}

	newExpectedJSON := `{"sweetCodes":""}`
	if string(encoded) != newExpectedJSON {
		t.Errorf("Currency Code Upper() error expected %s but received %s",
			newExpectedJSON, string(encoded))
	}
}

func TestIsFiatCurrency(t *testing.T) {
	if (Code{}).IsFiatCurrency() {
		t.Errorf("TestIsFiatCurrency cannot match currency, %s.",
			Code{})
	}
	if !USD.IsFiatCurrency() {
		t.Errorf(
			"TestIsFiatCurrency cannot match currency, %s.", USD)
	}
	if !CNY.IsFiatCurrency() {
		t.Errorf(
			"TestIsFiatCurrency cannot match currency, %s.", CNY)
	}
	if LINO.IsFiatCurrency() {
		t.Errorf(
			"TestIsFiatCurrency cannot match currency, %s.", LINO,
		)
	}
	if USDT.IsFiatCurrency() {
		t.Errorf(
			"TestIsFiatCurrency cannot match currency, %s.", USD)
	}
	if DAI.IsFiatCurrency() {
		t.Errorf(
			"TestIsFiatCurrency cannot match currency, %s.", USD)
	}
}

func TestIsCryptocurrency(t *testing.T) {
	if (Code{}).IsCryptocurrency() {
		t.Errorf("TestIsCryptocurrency cannot match currency, %s.",
			Code{})
	}
	if !BTC.IsCryptocurrency() {
		t.Errorf("TestIsCryptocurrency cannot match currency, %s.",
			BTC)
	}
	if !LTC.IsCryptocurrency() {
		t.Errorf("TestIsCryptocurrency cannot match currency, %s.",
			LTC)
	}
	if AUD.IsCryptocurrency() {
		t.Errorf("TestIsCryptocurrency cannot match currency, %s.",
			AUD)
	}
	if !USDT.IsCryptocurrency() {
		t.Errorf(
			"TestIsCryptocurrency cannot match currency, %s.", USD)
	}
	if !DAI.IsCryptocurrency() {
		t.Errorf(
			"TestIsCryptocurrency cannot match currency, %s.", USD)
	}
}

func TestIsStableCurrency(t *testing.T) {
	if (Code{}).IsStableCurrency() {
		t.Errorf("TestIsStableCurrency cannot match currency, %s.", Code{})
	}
	if BTC.IsStableCurrency() {
		t.Errorf("TestIsStableCurrency cannot match currency, %s.", BTC)
	}
	if LTC.IsStableCurrency() {
		t.Errorf("TestIsStableCurrency cannot match currency, %s.", LTC)
	}
	if AUD.IsStableCurrency() {
		t.Errorf("TestIsStableCurrency cannot match currency, %s.", AUD)
	}
	if !USDT.IsStableCurrency() {
		t.Errorf("TestIsStableCurrency cannot match currency, %s.", USDT)
	}
	if !DAI.IsStableCurrency() {
		t.Errorf("TestIsStableCurrency cannot match currency, %s.", DAI)
	}
}

func TestItemString(t *testing.T) {
	newItem := Item{
		ID:         1337,
		FullName:   "Hello,World",
		Symbol:     "HWORLD",
		AssocChain: "Silly",
	}

	if expected := "HWORLD"; newItem.String() != expected {
		t.Errorf("Currency String() error expected '%s' but received '%s'",
			expected,
			&newItem)
	}
}
