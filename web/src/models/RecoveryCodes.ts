// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

export interface RecoveryCodesStatus {
    codes_remaining: number;
    codes_total: number;
    generated_at?: Date;
    last_used_at?: Date;
}

export interface RecoveryCodesStatusPayload {
    codes_remaining: number;
    codes_total: number;
    generated_at?: string;
    last_used_at?: string;
}

export interface GenerateRecoveryCodesResponse {
    codes: string[];
    notification_sent: boolean;
}
