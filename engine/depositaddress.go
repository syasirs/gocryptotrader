package engine

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/deposit"
	"github.com/thrasher-corp/gocryptotrader/log"
)

const (
	// depositAddressDefaultMaxRetries max default retries when deposit address
	// is being created.
	depositAddressDefaultMaxRetries = 6
	// depositAddressDefaultWait defines the sleep time before another request
	// is sent.
	depositAddressDefaultWait = time.Second * 10
)

// vars related to the deposit address helpers
var (
	ErrDepositAddressStoreIsNil    = errors.New("deposit address store is nil")
	ErrDepositAddressNotFound      = errors.New("deposit address does not exist")
	ErrNoDepositAddressesRetrieved = errors.New("no deposit addresses retrieved")

	errDepositAddressChainNotFound = errors.New("deposit address for specified chain not found")
	errIsNotCryptocurrency         = errors.New("currency is not a cryptocurrency")
	errDepositAddressNotGenerated  = errors.New("deposit address not generated by exchange within required timeframe")
)

// DepositAddressManager manages the exchange deposit address store
type DepositAddressManager struct {
	m     sync.RWMutex
	store map[string]map[*currency.Item][]deposit.Address
}

// IsSynced returns whether or not the deposit address store has synced its data
func (m *DepositAddressManager) IsSynced() bool {
	if m.store == nil {
		return false
	}
	m.m.RLock()
	defer m.m.RUnlock()
	return len(m.store) > 0
}

// SetupDepositAddressManager returns a DepositAddressManager
func SetupDepositAddressManager() *DepositAddressManager {
	return &DepositAddressManager{
		store: make(map[string]map[*currency.Item][]deposit.Address),
	}
}

// GetDepositAddressByExchangeAndCurrency returns a deposit address for the specified exchange and cryptocurrency
// if it exists
func (m *DepositAddressManager) GetDepositAddressByExchangeAndCurrency(exchName, chain string, cc currency.Code) (deposit.Address, error) {
	if exchName == "" {
		return deposit.Address{}, errExchangeNameIsEmpty
	}

	if cc.IsEmpty() {
		return deposit.Address{}, fmt.Errorf("%s %s %w", exchName, chain, currency.ErrCurrencyCodeEmpty)
	}

	if cc.IsFiatCurrency() {
		return deposit.Address{}, fmt.Errorf("%s %s %s %w", exchName, chain, cc, errIsNotCryptocurrency)
	}

	m.m.RLock()
	defer m.m.RUnlock()
	if len(m.store) == 0 {
		return deposit.Address{}, ErrDepositAddressStoreIsNil
	}

	r, ok := m.store[strings.ToUpper(exchName)]
	if !ok {
		return deposit.Address{}, ErrExchangeNotFound
	}

	addresses, ok := r[cc.Item]
	if !ok {
		return deposit.Address{}, ErrDepositAddressNotFound
	}

	if len(addresses) == 0 {
		return deposit.Address{}, ErrNoDepositAddressesRetrieved
	}

	if chain != "" {
		for x := range addresses {
			if strings.EqualFold(addresses[x].Chain, chain) {
				return addresses[x], nil
			}
		}
		return deposit.Address{}, errDepositAddressChainNotFound
	}

	for x := range addresses {
		if strings.EqualFold(addresses[x].Chain, cc.String()) {
			return addresses[x], nil
		}
	}
	return addresses[0], nil
}

// GetDepositAddressByExchangeAndCurrency returns all deposit addresses and
// chains for a specific cryptocurrency.
func (m *DepositAddressManager) GetDepositAddressesByExchangeAndCurrency(exchName string, cc currency.Code) ([]deposit.Address, error) {
	if exchName == "" {
		return nil, errExchangeNameIsEmpty
	}

	if cc.IsEmpty() {
		return nil, fmt.Errorf("%s %w", exchName, currency.ErrCurrencyCodeEmpty)
	}

	if cc.IsFiatCurrency() {
		return nil, fmt.Errorf("%s %s %w", exchName, cc, errIsNotCryptocurrency)
	}

	m.m.RLock()
	defer m.m.RUnlock()
	if len(m.store) == 0 {
		return nil, ErrDepositAddressStoreIsNil
	}

	r, ok := m.store[strings.ToUpper(exchName)]
	if !ok {
		return nil, ErrExchangeNotFound
	}

	addresses, ok := r[cc.Item]
	if !ok {
		return nil, ErrDepositAddressNotFound
	}
	if len(addresses) == 0 {
		return nil, ErrNoDepositAddressesRetrieved
	}
	return addresses, nil
}

// GetDepositAddressesByExchange returns a list of cryptocurrency addresses for the specified
// exchange if they exist
func (m *DepositAddressManager) GetDepositAddressesByExchange(exchName string) (map[currency.Code][]deposit.Address, error) {
	m.m.RLock()
	defer m.m.RUnlock()

	if len(m.store) == 0 {
		return nil, ErrDepositAddressStoreIsNil
	}

	r, ok := m.store[strings.ToUpper(exchName)]
	if !ok {
		return nil, ErrDepositAddressNotFound
	}

	cpy := make(map[currency.Code][]deposit.Address, len(r))
	for item, addresses := range r {
		addrsCpy := make([]deposit.Address, len(addresses))
		copy(addrsCpy, addresses)
		cpy[currency.Code{Item: item}] = addrsCpy
	}
	return cpy, nil
}

// Sync synchronises all deposit addresses
func (m *DepositAddressManager) Sync(addresses map[string]map[currency.Code][]deposit.Address) error {
	if m == nil {
		return fmt.Errorf("deposit address manager %w", ErrNilSubsystem)
	}

	if len(addresses) == 0 {
		return fmt.Errorf("deposit address manager %w", ErrNoDepositAddressesRetrieved)
	}

	m.m.Lock()
	defer m.m.Unlock()
	if m.store == nil {
		return ErrDepositAddressStoreIsNil
	}
	for exchName, m1 := range addresses {
		r := make(map[*currency.Item][]deposit.Address)
		for code, addresses := range m1 {
			r[code.Item] = addresses
		}
		m.store[strings.ToUpper(exchName)] = r
	}
	return nil
}

// FetchDepositAddressWithRetry fetches a deposit address from an exchange on
// an exchange generating an address will retry with params.
func FetchDepositAddressWithRetry(ctx context.Context, exch exchange.IBotExchange, cc currency.Code, chain string, retryCount int, sleepTime time.Duration) (*deposit.Address, error) {
	if retryCount <= 0 {
		retryCount = depositAddressDefaultMaxRetries
	}
	if sleepTime <= 0 {
		sleepTime = depositAddressDefaultWait
	}
	for x := 0; x < retryCount; x++ {
		address, err := exch.GetDepositAddress(ctx, cc, "", chain)
		if err == nil {
			return address, nil
		}

		if !errors.Is(err, deposit.ErrAddressBeingCreated) {
			return nil, fmt.Errorf("%s failed to get cryptocurrency deposit address for %s. Err: %w",
				exch.GetName(),
				cc,
				err)
		}

		log.Warnf(log.ExchangeSys,
			"Deposit address for %s %s %s is being generated, sleeping: %s count: %d",
			exch.GetName(), cc, chain, sleepTime, x+1)
		time.Sleep(sleepTime)
	}
	return nil, fmt.Errorf("%s %s %s %w SleepTime: %s Retries: %d",
		exch.GetName(),
		cc,
		chain,
		errDepositAddressNotGenerated,
		sleepTime,
		retryCount)
}
