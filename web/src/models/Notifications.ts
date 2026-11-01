// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

export type NotificationLevel = "error" | "info" | "success" | "warning";

export interface Notification {
    message: string;
    level: NotificationLevel;
    timeout: number;
}
