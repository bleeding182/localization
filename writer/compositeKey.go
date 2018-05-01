package writer

import "regexp"

// We support keys in the form of [group]__?[identifier]__pl_[<one|other|etc>]
// For multi-word groups we can use __ instead to create a long_group__identifier
// 0: original key
// 1: key without (optional) quanity
// 2: group part of key (part until __ or first part until _)
// 3: identifier - part without group or quantity
// 4: optional quantity
var keyRegex = regexp.MustCompile("^((.*?)(?:_{1,2})((?:[a-zA-Z0-9]+_)*[a-zA-Z0-9]*))(?:__pl_(.*))?$")

// CompositeKeyOf returns a new CompositeKey after parsing the key argument.
func CompositeKeyOf(key string) CompositeKey {
	return CompositeKey{keyRegex.FindStringSubmatch(key)}
}

// CompositeKey represents a key that consists of [group]_[identifier]_[plural]
type CompositeKey struct {
	parts []string
}

// Original the complete, original key
func (key CompositeKey) Original() string {
	return key.parts[0]
}

// PlainKey the full key without any (optional) quantity
func (key CompositeKey) PlainKey() string {
	return key.parts[1]
}

// Group of the key, the first part before any `_` or `__` for longer names
func (key CompositeKey) Group() string {
	return key.parts[2]
}

// Identifier of the Key without a group or (optional) quantity
func (key CompositeKey) Identifier() string {
	return key.parts[3]
}

// Quantity of the key (optional)
func (key CompositeKey) Quantity() string {
	return key.parts[4]
}
