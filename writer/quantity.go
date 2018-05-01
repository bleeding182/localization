package writer

import "fmt"

// Quantity modifier of a pluralizable string.
type Quantity int

// QuantityString combines multiple variant of a string. e.g. [one] book [other] books
type QuantityString struct {
	Key    string
	Values map[Quantity]LocalizedString
}

// https://developer.android.com/guide/topics/resources/string-resource.html#Plurals
const (
	zero  Quantity = 0 // When the language requires special treatment of the number 0 (as in Arabic).
	one   Quantity = 1 // When the language requires special treatment of numbers like one (as with the number 1 in English and most other languages; in Russian, any number ending in 1 but not ending in 11 is in this class).
	two   Quantity = 2 // When the language requires special treatment of numbers like two (as with 2 in Welsh, or 102 in Slovenian).
	few   Quantity = 3 // When the language requires special treatment of "small" numbers (as with 2, 3, and 4 in Czech; or numbers ending 2, 3, or 4 but not 12, 13, or 14 in Polish).
	many  Quantity = 4 // When the language requires special treatment of "large" numbers (as with numbers ending 11-99 in Maltese).
	other Quantity = 5 // When the language does not require special treatment of the given quantity (as with all numbers in Chinese, or 42 in English).
)

var quantities = map[Quantity]string{
	zero:  "zero",
	one:   "one",
	two:   "two",
	few:   "few",
	many:  "many",
	other: "other",
}

// QuantityOf the quantityString mapped to Quantity, e.g. "zero" -> 0
func QuantityOf(quantityString string) Quantity {
	for q, s := range quantities {
		if s == quantityString {
			return q
		}
	}
	panic(fmt.Sprint("Unknown quantity", quantityString))
}

func (quantity Quantity) String() string {
	if quantity < zero || quantity > other {
		panic(fmt.Sprint("Unknown quantity", uint(quantity)))
	}
	return quantities[quantity]
}
