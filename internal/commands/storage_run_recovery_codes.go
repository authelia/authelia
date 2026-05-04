// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package commands

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/random"
)

type cmdRecoveryCodeContext struct {
	context.Context
	clk clock.Provider
	rnd random.Provider
}

func (c cmdRecoveryCodeContext) GetClock() clock.Provider { return c.clk }

func (c cmdRecoveryCodeContext) GetRandom() random.Provider { return c.rnd }

func (c cmdRecoveryCodeContext) RemoteIP() net.IP { return net.IP{} }

// StorageUserRecoveryCodesStatusRunE is the RunE for the authelia storage user recovery-codes status command.
func (ctx *CmdCtx) StorageUserRecoveryCodesStatusRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	user := args[0]

	codes, err := ctx.providers.StorageProvider.LoadRecoveryCodesByUsername(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to load recovery codes for user '%s': %w", user, err)
	}

	var unused, consumed, revoked int

	for _, c := range codes {
		switch {
		case c.RevokedAt.Valid:
			revoked++
		case c.ConsumedAt.Valid:
			consumed++
		default:
			unused++
		}
	}

	w := cmd.OutOrStdout()

	_, _ = fmt.Fprintf(w, "User: %s\n", user)
	_, _ = fmt.Fprintf(w, "Recovery codes total: %d (unused: %d, consumed: %d, revoked: %d)\n", len(codes), unused, consumed, revoked)

	return nil
}

// StorageUserRecoveryCodesListRunE is the RunE for the authelia storage user recovery-codes list command.
func (ctx *CmdCtx) StorageUserRecoveryCodesListRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	user := args[0]

	codes, err := ctx.providers.StorageProvider.LoadRecoveryCodesByUsername(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to load recovery codes for user '%s': %w", user, err)
	}

	w := cmd.OutOrStdout()

	if len(codes) == 0 {
		_, _ = fmt.Fprintf(w, "User '%s' has no recovery codes.\n", user)

		return nil
	}

	_, _ = fmt.Fprintf(w, "ID\tCreated\tStatus\n")

	for _, c := range codes {
		status := "unused"

		switch {
		case c.RevokedAt.Valid:
			status = fmt.Sprintf("revoked at %s", c.RevokedAt.Time.Format("2006-01-02 15:04 MST"))
		case c.ConsumedAt.Valid:
			status = fmt.Sprintf("consumed at %s", c.ConsumedAt.Time.Format("2006-01-02 15:04 MST"))
		}

		_, _ = fmt.Fprintf(w, "%d\t%s\t%s\n", c.ID, c.CreatedAt.Format("2006-01-02 15:04 MST"), status)
	}

	return nil
}

// StorageUserRecoveryCodesGenerateRunE is the RunE for the authelia storage user recovery-codes generate command.
func (ctx *CmdCtx) StorageUserRecoveryCodesGenerateRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	user := args[0]

	if err = ctx.providers.StorageProvider.RevokeRecoveryCodesByUsername(ctx, user, model.NullIP{}); err != nil {
		return fmt.Errorf("failed to revoke existing recovery codes for user '%s': %w", user, err)
	}

	mc := cmdRecoveryCodeContext{Context: ctx, clk: ctx.GetClock(), rnd: ctx.GetRandom()}

	codes := make([]string, 0, model.RecoveryCodeBatchSize)

	for i := 0; i < model.RecoveryCodeBatchSize; i++ {
		code, errGen := model.NewRecoveryCode(mc, user)
		if errGen != nil {
			return fmt.Errorf("failed to generate recovery code %d for user '%s': %w", i, user, errGen)
		}

		if err = ctx.providers.StorageProvider.SaveRecoveryCode(ctx, code); err != nil {
			return fmt.Errorf("failed to save recovery code %d for user '%s': %w", i, user, err)
		}

		codes = append(codes, code.Plaintext)
	}

	w := cmd.OutOrStdout()

	_, _ = fmt.Fprintf(w, "Generated %d recovery codes for user '%s'. Deliver these to the user; they will not be shown again:\n\n", len(codes), user)
	_, _ = fmt.Fprintln(w, strings.Join(codes, "\n"))

	return nil
}

// StorageUserRecoveryCodesDeleteRunE is the RunE for the authelia storage user recovery-codes delete command.
func (ctx *CmdCtx) StorageUserRecoveryCodesDeleteRunE(cmd *cobra.Command, args []string) (err error) {
	defer func() {
		if err := ctx.providers.StorageProvider.Close(); err != nil {
			panic(err)
		}
	}()

	if err = ctx.CheckSchema(); err != nil {
		return storageWrapCheckSchemaErr(err)
	}

	user := args[0]

	if err = ctx.providers.StorageProvider.RevokeRecoveryCodesByUsername(ctx, user, model.NullIP{}); err != nil {
		return fmt.Errorf("failed to revoke recovery codes for user '%s': %w", user, err)
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Successfully revoked all recovery codes for user '%s'.\n", user)

	return nil
}
