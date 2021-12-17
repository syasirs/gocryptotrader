package bank

import (
	"errors"
	"fmt"
)

// Custom types for different international banking options
const (
	NotApplicable Transfer = iota + 1
	WireTransfer
	ExpressWireTransfer
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
	FasterPaymentService
	MobileMoney
	CashTransfer
	YandexMoney
	GEOPay
	SettlePay
	ExchangeFiatDWChannelSignetUSD         // Binance
	ExchangeFiatDWChannelSwiftSignatureBar // Binance
	AutomaticClearingHouse
	FedWire
	TelegraphicTransfer // Coinut
	SDDomesticCheque    // Coinut
	Xfers               // Coinut
	ExmoGiftCard        // Exmo
	Terminal            // Exmo
)

var (
	// ErrUnknownTransfer defines an unknown bank transfer type error
	ErrUnknownTransfer = errors.New("unknown bank transfer type")
	// ErrTransferTypeUnset defines an error when the transfer type is unset
	ErrTransferTypeUnset = errors.New("transfer type is unset")
)

// Transfer defines the different fee types associated with bank
// transactions to and from an exchange.
type Transfer uint8

// String implements the stringer interface
func (b Transfer) String() string {
	switch b {
	case NotApplicable:
		return "NotApplicable"
	case WireTransfer:
		return "WireTransfer"
	case ExpressWireTransfer:
		return "ExpressWireTransfer"
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
	case FasterPaymentService:
		return "FasterPaymentService(FPS)"
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
	case AutomaticClearingHouse:
		return "AutomaticClearingHouse"
	case FedWire:
		return "FedWire"
	case TelegraphicTransfer:
		return "TelegraphicTransfer"
	case SDDomesticCheque:
		return "SDDomesticCheque"
	case Xfers:
		return "Xfers"
	case ExmoGiftCard:
		return "ExmoGiftCard"
	case Terminal:
		return "Terminal"
	default:
		return ""
	}
}

// Validate validates an international bank transaction option
func (b Transfer) Validate() error {
	switch b {
	case 0:
		return ErrTransferTypeUnset
	case NotApplicable,
		WireTransfer,
		ExpressWireTransfer,
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
		FasterPaymentService,
		MobileMoney,
		CashTransfer,
		YandexMoney,
		GEOPay,
		SettlePay,
		ExchangeFiatDWChannelSignetUSD,
		ExchangeFiatDWChannelSwiftSignatureBar,
		AutomaticClearingHouse,
		FedWire,
		TelegraphicTransfer,
		SDDomesticCheque,
		Xfers,
		ExmoGiftCard,
		Terminal:
		return nil
	default:
		return fmt.Errorf("%d: %w", b, ErrUnknownTransfer)
	}
}
