package domain

import "github.com/shopspring/decimal"

type ProductSort string

const (
	ProductSortPriceAsc  ProductSort = "price_asc"
	ProductSortPriceDesc ProductSort = "price_desc"
	ProductSortNameAsc   ProductSort = "name_asc"
	ProductSortNameDesc  ProductSort = "name_desc"
	ProductSortNewest    ProductSort = "newest"
)

// ParseProductSort validates raw against the known ProductSort enum values.
// An empty raw is not a valid sort on its own — callers treat "sort not sent"
// as a zero ProductFilter.Sort before calling this.
func ParseProductSort(raw string) (ProductSort, bool) {
	switch ProductSort(raw) {
	case ProductSortPriceAsc, ProductSortPriceDesc, ProductSortNameAsc, ProductSortNameDesc, ProductSortNewest:
		return ProductSort(raw), true
	default:
		return "", false
	}
}

// ProductFilter narrows a Product listing. Query matches only name/sku/description
// (category is matched exclusively via Category — AYD-006 narrowing). PriceMin/
// PriceMax are nil when not sent.
type ProductFilter struct {
	Query    string
	Category string
	PriceMin *decimal.Decimal
	PriceMax *decimal.Decimal
	Sort     ProductSort
}

// Validate reports problems already-parsed values carry: a negative bound, an
// inverted range, or a sort value outside the enum. Decimal parse failures are
// caught by the caller (the handler) before a ProductFilter is even built and
// use the same "must_be_non_negative_decimal" detail code.
func (f ProductFilter) Validate() map[string]string {
	problems := map[string]string{}

	if f.PriceMin != nil && f.PriceMin.IsNegative() {
		problems["price_min"] = "must_be_non_negative_decimal"
	}
	if f.PriceMax != nil && f.PriceMax.IsNegative() {
		problems["price_max"] = "must_be_non_negative_decimal"
	}

	if _, alreadyInvalid := problems["price_min"]; !alreadyInvalid {
		if f.PriceMin != nil && f.PriceMax != nil && f.PriceMin.GreaterThan(*f.PriceMax) {
			problems["price_min"] = "must_not_exceed_price_max"
		}
	}

	if f.Sort != "" {
		if _, ok := ParseProductSort(string(f.Sort)); !ok {
			problems["sort"] = "invalid_sort"
		}
	}

	return problems
}
