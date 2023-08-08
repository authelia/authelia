package model

import "github.com/google/uuid"

func NewRandomNullUUID() (uuid.NullUUID, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return uuid.NullUUID{}, err
	}

	return uuid.NullUUID{UUID: id, Valid: true}, nil
}

func NullUUID(in uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: in, Valid: in.ID() != 0}
}

func MustNullUUID(in uuid.NullUUID, err error) uuid.NullUUID {
	if err != nil {
		panic(err)
	}

	return in
}
