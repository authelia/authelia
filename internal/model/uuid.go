package model

import "github.com/google/uuid"

// NewRandomNullUUID returns a uuid.NullUUID using the uud.NewRandom() method i.e. in the form of a v4 UUID.
func NewRandomNullUUID() (uuid.NullUUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.NullUUID{}, err
	}

	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

// NullUUID converts a uuid.UUID to a uuid.NullUUID.
func NullUUID(in uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: in, Valid: in != uuid.Nil}
}

// MustNullUUID is a uuid.Must variant for the uuid.NullUUID methods.
func MustNullUUID(in uuid.NullUUID, err error) uuid.NullUUID {
	if err != nil {
		panic(err)
	}

	return in
}
