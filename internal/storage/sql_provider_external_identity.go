// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/model"
)

// externalIdentityLinkSignatureContext is signed ahead of the values of a link, so a signature made for any other
// purpose with the same key can never be mistaken for the signature of a link, and so the values signed can be
// changed in the future without a signature of one form being valid for the other.
const externalIdentityLinkSignatureContext = "authelia:external_identity_link:v1"

// externalIdentityLinkSignature returns the HMAC signature of every security sensitive value of a link: the provider type,
// the provider, the issuer, and the subject which together identify the external identity, and the username of the account it
// signs in to. Each value is written with its length ahead of it so that no two different sets of values can be
// encoded the same way, e.g. an issuer ending with the start of the subject.
func externalIdentityLinkSignature(key []byte, link *model.ExternalIdentityLink) string {
	h := hmac.New(sha512.New, key)

	var length [8]byte

	for _, value := range []string{externalIdentityLinkSignatureContext, link.Type, link.Provider, link.Issuer, link.Subject, link.Username} {
		binary.BigEndian.PutUint64(length[:], uint64(len(value)))

		h.Write(length[:])
		h.Write([]byte(value))
	}

	return hex.EncodeToString(h.Sum(nil))
}

func (p *SQLProvider) externalIdentityLinkHMACSignature(link *model.ExternalIdentityLink) string {
	return externalIdentityLinkSignature(p.keys.externalIdentityLinkHMAC, link)
}

// verifyExternalIdentityLink returns an error when the signature of the link is not the signature of its values.
func (p *SQLProvider) verifyExternalIdentityLink(link *model.ExternalIdentityLink) (err error) {
	if !verifyExternalIdentityLinkSignature(p.keys.externalIdentityLinkHMAC, link) {
		return fmt.Errorf("error verifying external identity link with id '%d': %w", link.ID, ErrExternalIdentityLinkSignatureInvalid)
	}

	return nil
}

func verifyExternalIdentityLinkSignature(key []byte, link *model.ExternalIdentityLink) bool {
	if len(key) == 0 || link.Signature == "" {
		return false
	}

	return hmac.Equal([]byte(externalIdentityLinkSignature(key, link)), []byte(link.Signature))
}

func (p *SQLProvider) getHMACExternalIdentityLink(ctx context.Context) (key []byte, err error) {
	return p.getHMACKey(ctx, hmacNameExternalIdentityLink, sha512.BlockSize)
}

// schemaEncryptionRotateHMACKeyExternalIdentityLink rotates the external identity link HMAC key. Unlike the other HMAC
// keys the table it protects is not truncated, as that would remove every user's links: every link with a signature
// which is valid for the current key is signed again with the new key, and every link without one is deleted, as it is
// not a link Authelia wrote and signing it again would make it one.
func (p *SQLProvider) schemaEncryptionRotateHMACKeyExternalIdentityLink(ctx context.Context) (err error) {
	var tx SQLXTx

	if tx, err = p.db.Beginx(); err != nil {
		return fmt.Errorf("error beginning transaction to rotate hmac key: %w", err)
	}

	var (
		current, key []byte
		signed       int
		deleted      int
	)

	rollback := func(err error) error {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return fmt.Errorf("error rolling back transaction to rotate hmac key: %w", rollbackErr)
		}

		return err
	}

	if current, err = p.getEncryptionValue(ctx, tx, fmt.Sprintf(fmtNameKeyHMAC, hmacNameExternalIdentityLink)); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return rollback(fmt.Errorf("error getting the current hmac key: %w", err))
	}

	var links []model.ExternalIdentityLink

	if err = tx.SelectContext(ctx, &links, p.db.Rebind(fmt.Sprintf(queryFmtSelectExternalIdentityLinks, tableUserExternalIdentityLinks))); err != nil {
		return rollback(fmt.Errorf("error selecting external identity links: %w", err))
	}

	if key, err = p.setCrypographyKey(ctx, tx, keyTypeCryptographyHMAC, hmacNameExternalIdentityLink, sha512.BlockSize, true); err != nil {
		return rollback(fmt.Errorf("error setting the hmac key: %w", err))
	}

	queryUpdate := p.db.Rebind(fmt.Sprintf(queryFmtUpdateExternalIdentityLinkSignature, tableUserExternalIdentityLinks))
	queryDelete := p.db.Rebind(fmt.Sprintf(queryFmtDeleteExternalIdentityLinkByID, tableUserExternalIdentityLinks))

	for i := range links {
		link := &links[i]

		if !verifyExternalIdentityLinkSignature(current, link) {
			p.log.WithField("id", link.ID).Warn("Deleting an external identity link during the hmac key rotation as its signature is not valid for the current key")

			if _, err = tx.ExecContext(ctx, queryDelete, link.ID); err != nil {
				return rollback(fmt.Errorf("error deleting external identity link with id '%d': %w", link.ID, err))
			}

			deleted++

			continue
		}

		if _, err = tx.ExecContext(ctx, queryUpdate, externalIdentityLinkSignature(key, link), link.ID); err != nil {
			return rollback(fmt.Errorf("error signing external identity link with id '%d': %w", link.ID, err))
		}

		signed++
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("error committing transaction to rotate hmac key: %w", err)
	}

	p.keys.externalIdentityLinkHMAC = key

	p.log.Infof("Finished signing %d external identity link(s) with the new hmac key and deleted %d link(s) with a signature which was not valid", signed, deleted)

	return nil
}
