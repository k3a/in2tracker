package marketdata

import "strings"

// Market holds market identifier
type Market struct {
	codeID      uint16
	idents      []string
	receiverMap map[string]uint16
}

func (mkt *Market) String() string {
	if mkt == nil {
		return "(nil)"
	}
	return strings.Join(mkt.idents, ",")
}

// IdentifierForReceiver returns identifier used by the specified receiver
// or empty string if there is no binding
func (mkt *Market) IdentifierForReceiver(receiver string) string {
	if mkt.receiverMap == nil {
		return ""
	}

	u, has := mkt.receiverMap[receiver]
	if !has {
		return ""
	}

	if int(u) >= len(mkt.idents) {
		return ""
	}

	return mkt.idents[u]
}

// findIdentIndex returns index in idents array and bool telling
// whether the identifier was found in the array or not
func (mkt *Market) findIdentIndex(ident string) (uint16, bool) {
	for i, name := range mkt.idents {
		if name == ident {
			return uint16(i), true
		}
	}

	return 0, false
}

// AssignIdentifierForReceiver assigns an identifier for the market used by the named receiver
func (mkt *Market) AssignIdentifierForReceiver(identifier, receiver string) {
	if mkt.receiverMap == nil {
		mkt.receiverMap = make(map[string]uint16)
	}

	// try to find existing identifier id
	idx, found := mkt.findIdentIndex(identifier)
	if found {
		mkt.receiverMap[receiver] = idx
		return
	}

	mkt.receiverMap[receiver] = uint16(len(mkt.idents))
	mkt.idents = append(mkt.idents, identifier)
}

var markets []*Market

func registerMarket(idents ...string) *Market {
	m := &Market{uint16(len(markets)), idents, nil}

	markets = append(markets, m)

	return m
}

// Market identifier
var (
	MarketAny                    = registerMarket("")
	MarketUSANYSE                = registerMarket("NYSE", "XNYS", "NYQ")
	MarketUSANYSEArca            = registerMarket("NYSEARCA", "ARCX")
	MarketUSANasdaq              = registerMarket("NASDAQ", "XNGS", "XNAS", "NMS")
	MarketsEuropeFrankfurtBoerse = registerMarket("FWB", "FRA")
	MarketsEuropeFrankfurtXETRA  = registerMarket("IBIS", "XETRA")
	MarketsEuropeLSE             = registerMarket("LSE", "TRQXUK")
)

// MarketFromString returns market identifier from string or MarketAny if not known
func MarketFromString(ident string) *Market {
	if len(ident) == 0 {
		return MarketAny
	}

	for _, m := range markets {
		for _, mid := range m.idents {
			if strings.EqualFold(ident, mid) {
				return m
			}
		}
	}
	return MarketAny
}

// MarketFromStringForReceiver returns market known by the receiver under the provided identifier.
// Additional receivers are registered for markets using Market.AssignIdentifierForReceiver.
// Returns MarketAny if the identifier is not known
func MarketFromStringForReceiver(ident, receiver string) *Market {
	if len(ident) == 0 {
		return MarketAny
	}

	for _, m := range markets {
		if m.IdentifierForReceiver(receiver) == ident {
			return m
		}
	}
	return MarketAny
}

// MarketEquals returns true if the both market types represend the same market
func MarketEquals(a *Market, b *Market) bool { //nolint:gocyclo
	// equals only if: both nil, nil == any, any == nil
	if a == nil && b == nil {
		return true
	} else if a == nil && b != nil && b.codeID == MarketAny.codeID {
		return true
	} else if a != nil && b == nil && a.codeID == MarketAny.codeID {
		return true
	} else if a == nil || b == nil {
		return false
	}

	// or if both code identifiers are the same
	return a.codeID == b.codeID
}
