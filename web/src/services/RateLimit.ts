// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import axios, { AxiosInstance, AxiosResponse } from "axios";

export const DefaultRetryAfterSeconds = 10;
export const MinimumRetryAfterSeconds = 1;
export const MaximumRetryAfterSeconds = 86400;

export class RateLimitedError extends Error {
    readonly retryAfter: number;

    constructor(retryAfter: number) {
        super("Too Many Requests");
        this.name = "RateLimitedError";
        this.retryAfter = retryAfter;
    }
}

export function getRetryAfter(resp: AxiosResponse): number {
    const value = resp.headers?.["retry-after"];

    if (typeof value !== "string" || value === "") {
        return DefaultRetryAfterSeconds;
    }

    if (/^\d+$/.test(value)) {
        return clampRetryAfter(Number.parseInt(value, 10));
    }

    const date = new Date(value);

    if (Number.isNaN(date.getTime())) {
        return DefaultRetryAfterSeconds;
    }

    return clampRetryAfter(Math.ceil((date.getTime() - Date.now()) / 1000));
}

function clampRetryAfter(seconds: number): number {
    if (!Number.isFinite(seconds)) {
        return DefaultRetryAfterSeconds;
    }

    return Math.min(MaximumRetryAfterSeconds, Math.max(MinimumRetryAfterSeconds, seconds));
}

export function registerRateLimitInterceptor(instance: AxiosInstance = axios) {
    return instance.interceptors.response.use(
        (resp) => {
            if (resp.status === 429) {
                throw new RateLimitedError(getRetryAfter(resp));
            }

            return resp;
        },
        (err) => {
            if (axios.isAxiosError(err) && err.response?.status === 429) {
                throw new RateLimitedError(getRetryAfter(err.response));
            }

            throw err;
        },
    );
}
