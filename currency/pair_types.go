package currency

// Pair holds currency pair information
type Pair struct {
	ID        string `json:"id"`
	Delimiter string `json:"delimiter,omitempty"`
	Base      Code   `json:"base,omitempty"`
	Quote     Code   `json:"quote,omitempty"`
	Index     string `json:"index,omitempty"`
}

// Pairs defines a list of pairs
type Pairs []*Pair
