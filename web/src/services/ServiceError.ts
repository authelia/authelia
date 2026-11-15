// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { ErrorCode } from "@services/ErrorCode";

export class ServiceError extends Error {
    readonly code?: ErrorCode;
    readonly status: number;

    constructor(message: string, status: number, code?: ErrorCode) {
        super(message);

        this.name = "ServiceError";
        this.status = status;
        this.code = code;
    }
}
