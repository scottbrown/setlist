package setlist

import (
	"errors"
	"strings"
)

var ErrEmptyString = errors.New("cannot be an empty string")

var ErrWrongFormat = errors.New("invalid format")

type IdentityStoreId string

type Region string

func NewIdentityStoreId(id string) (IdentityStoreId, error) {
	if id == "" {
		return IdentityStoreId(""), ErrEmptyString
	}

	if !strings.HasPrefix(id, "d-") {
		return IdentityStoreId(""), ErrWrongFormat
	}

	return IdentityStoreId(id), nil
}

func (i IdentityStoreId) String() string {
	return string(i)
}

func NewRegion(region string) (Region, error) {
	if region == "" {
		return Region(""), ErrEmptyString
	}

	return Region(region), nil
}

func (r Region) String() string {
	return string(r)
}
