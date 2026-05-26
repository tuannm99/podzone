package entity

import "errors"

const MembershipStatusActive = "active"

var (
	ErrMembershipNotFound = errors.New("tenant membership not found")
	ErrInactiveMembership = errors.New("tenant membership is not active")
)
