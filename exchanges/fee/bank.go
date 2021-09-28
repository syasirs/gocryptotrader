package fee

import (
	"errors"
	"fmt"
)

// Custom types for different internation banking options
const (
	NotApplicable BankTransaction = iota
	WireTransfer
	PerfectMoney
	Neteller
	AdvCash
	Payeer
	Skrill
	Simplex
	SEPA
	Swift
	RapidTransfer
	MisterTangoSEPA
	Qiwi
	VisaMastercard
	WebMoney
	Capitalist
	WesternUnion
	MoneyGram
	Contact
	PayIDOsko
	BankCardVisa
	BankCardMastercard
	BankCardMIR // Russian credit card provider
	CreditCardMastercard
	Sofort
	P2P
	Etana
	FasterPayments
	MobileMoney
	CashTransfer
	YandexMoney
	GEOPay
	SettlePay
	ExchangeFiatDWChannelSignetUSD         // Binance
	ExchangeFiatDWChannelSwiftSignatureBar // Binance
)

var errUnknownBankTransaction = errors.New("unknown bank transaction type")

// BankTransaction defines the different fee types associated with bank
// transactions to and from an exchange.
type BankTransaction uint8

// String implements the stringer interface
func (b BankTransaction) String() string {
	switch b {
	case NotApplicable:
		return "NotApplicable"
	case WireTransfer:
		return "WireTransfer"
	case PerfectMoney:
		return "PerfectMoney"
	case Neteller:
		return "Neteller"
	case AdvCash:
		return "AdvCash"
	case Payeer:
		return "Payeer"
	case Skrill:
		return "Skrill"
	case Simplex:
		return "Simplex"
	case SEPA:
		return "SEPA"
	case Swift:
		return "Swift"
	case RapidTransfer:
		return "RapidTransfer"
	case MisterTangoSEPA:
		return "MisterTangoSEPA"
	case Qiwi:
		return "Qiwi"
	case VisaMastercard:
		return "VisaMastercard"
	case WebMoney:
		return "WebMoney"
	case Capitalist:
		return "Capitalist"
	case WesternUnion:
		return "WesternUnion"
	case MoneyGram:
		return "MoneyGram"
	case Contact:
		return "Contact"
	case PayIDOsko:
		return "PayID/Osko"
	case BankCardVisa:
		return "BankCard Visa"
	case BankCardMastercard:
		return "BankCard Mastercard"
	case BankCardMIR:
		return "BankCard MIR"
	case CreditCardMastercard:
		return "CreditCard Mastercard"
	case Sofort:
		return "Sofort"
	case P2P:
		return "P2P"
	case Etana:
		return "Etana"
	case FasterPayments:
		return "FasterPayments"
	case MobileMoney:
		return "MobileMoney"
	case CashTransfer:
		return "CashTransfer"
	case YandexMoney:
		return "YandexMoney"
	case GEOPay:
		return "GEOPay"
	case SettlePay:
		return "SettlePay"
	case ExchangeFiatDWChannelSignetUSD:
		return "ExchangeFiatDWChannelSignetUSD"
	case ExchangeFiatDWChannelSwiftSignatureBar:
		return "ExchangeFiatDWChannelSignetUSD"
	default:
		return ""
	}
}

// Validates an international bank transaction option
func (b BankTransaction) Validate() error {
	switch b {
	case NotApplicable,
		WireTransfer,
		PerfectMoney,
		Neteller,
		AdvCash,
		Payeer,
		Skrill,
		Simplex,
		SEPA,
		Swift,
		RapidTransfer,
		MisterTangoSEPA,
		Qiwi,
		VisaMastercard,
		WebMoney,
		Capitalist,
		WesternUnion,
		MoneyGram,
		Contact,
		PayIDOsko,
		BankCardVisa,
		BankCardMastercard,
		BankCardMIR,
		CreditCardMastercard,
		Sofort,
		P2P,
		Etana,
		FasterPayments,
		MobileMoney,
		CashTransfer,
		YandexMoney,
		GEOPay,
		SettlePay,
		ExchangeFiatDWChannelSignetUSD,
		ExchangeFiatDWChannelSwiftSignatureBar:
		return nil
	default:
		return fmt.Errorf("%d: %w", b, errUnknownBankTransaction)
	}
}
