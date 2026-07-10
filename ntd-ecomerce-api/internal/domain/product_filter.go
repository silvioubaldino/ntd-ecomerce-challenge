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

func ParseProductSort(raw string) (ProductSort, bool) {
	switch ProductSort(raw) {
	case ProductSortPriceAsc, ProductSortPriceDesc, ProductSortNameAsc, ProductSortNameDesc, ProductSortNewest:
		return ProductSort(raw), true
	default:
		return "", false
	}
}

type ProductFilter struct {
	Query    string
	Category string
	PriceMin *decimal.Decimal
	PriceMax *decimal.Decimal
	Sort     ProductSort
}

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
