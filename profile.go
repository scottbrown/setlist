package setlist

import (
	"fmt"
)

// Profile represents an AWS SSO profile configuration.
// It stores information about a permission set, including its metadata
// and the account it belongs to. This is used to generate AWS CLI profile
// configurations in the output file.
type Profile struct {
	Name            ProfileName
	Description     ProfileDescription
	SessionDuration SessionDuration
	SessionName     SessionName
	AccountId       AWSAccountId
	RoleName        RoleName
}

type ProfileName string
type ProfileDescription string
type SessionDuration string
type SessionName string
type AWSAccountId string
type RoleName string

func NewProfileName(name string) (ProfileName, error) {
	if name == "" {
		return ProfileName(""), ErrEmptyString
	}

	return ProfileName(name), nil
}

func NewProfileDescription(desc string) (ProfileDescription, error) {
	if desc == "" {
		return ProfileDescription(""), ErrEmptyString
	}

	return ProfileDescription(desc), nil
}

func NewSessionDuration(duration string) (SessionDuration, error) {
	if duration == "" {
		return SessionDuration(""), ErrEmptyString
	}

	return SessionDuration(duration), nil
}

func NewSessionName(name string) (SessionName, error) {
	if name == "" {
		return SessionName(""), ErrEmptyString
	}

	return SessionName(name), nil
}

var ErrInvalidAWSAccountIdLength = fmt.Errorf("invalid length for AWS account id")

func NewAWSAccountId(id string) (AWSAccountId, error) {
	if id == "" {
		return AWSAccountId(""), ErrEmptyString
	}

	if len(id) != 12 {
		return AWSAccountId(""), ErrInvalidAWSAccountIdLength
	}

	return AWSAccountId(id), nil
}

func NewRoleName(name string) (RoleName, error) {
	if name == "" {
		return RoleName(""), ErrEmptyString
	}

	return RoleName(name), nil
}

func (pn ProfileName) String() string {
	return string(pn)
}

func (pd ProfileDescription) String() string {
	return string(pd)
}

func (sd SessionDuration) String() string {
	return string(sd)
}

func (sn SessionName) String() string {
	return string(sn)
}

func (aai AWSAccountId) String() string {
	return string(aai)
}

func (rn RoleName) String() string {
	return string(rn)
}
