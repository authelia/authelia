// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import axios, { AxiosAdapter, AxiosError, AxiosResponse, InternalAxiosRequestConfig } from "axios";

import {
    DefaultRetryAfterSeconds,
    MaximumRetryAfterSeconds,
    MinimumRetryAfterSeconds,
    RateLimitedError,
    getRetryAfter,
    registerRateLimitInterceptor,
} from "@services/RateLimit";

function newResponse(status: number, headers: Record<string, string> = {}): AxiosResponse {
    return {
        config: { headers: {} } as InternalAxiosRequestConfig,
        data: { message: "Too Many Requests", status: "KO" },
        headers,
        status,
        statusText: "",
    };
}

function newInstance(status: number, headers: Record<string, string> = {}, validateStatus?: (s: number) => boolean) {
    const adapter: AxiosAdapter = async (config) => {
        const resp = { ...newResponse(status, headers), config };

        if (config.validateStatus && !config.validateStatus(status)) {
            throw new AxiosError("Request failed", undefined, config, undefined, resp);
        }

        return resp;
    };

    const instance = validateStatus ? axios.create({ adapter, validateStatus }) : axios.create({ adapter });

    registerRateLimitInterceptor(instance);

    return instance;
}

describe("getRetryAfter", () => {
    beforeEach(() => {
        vi.useFakeTimers();
        vi.setSystemTime(new Date("2026-01-01T00:00:00.000Z"));
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it.each([
        { expected: 30, headers: { "retry-after": "30" }, name: "delta seconds" },
        { expected: MinimumRetryAfterSeconds, headers: { "retry-after": "0" }, name: "zero delta seconds" },
        { expected: 60, headers: { "retry-after": "Thu, 01 Jan 2026 00:01:00 GMT" }, name: "http date" },
        {
            expected: MinimumRetryAfterSeconds,
            headers: { "retry-after": "Thu, 01 Jan 2026 00:00:00 GMT" },
            name: "http date now",
        },
        {
            expected: MinimumRetryAfterSeconds,
            headers: { "retry-after": "Wed, 31 Dec 2025 23:59:00 GMT" },
            name: "http date in the past",
        },
        {
            expected: MaximumRetryAfterSeconds,
            headers: { "retry-after": String(MaximumRetryAfterSeconds) },
            name: "maximum delta seconds",
        },
        {
            expected: MaximumRetryAfterSeconds,
            headers: { "retry-after": String(MaximumRetryAfterSeconds + 1) },
            name: "delta seconds above the maximum",
        },
        {
            expected: DefaultRetryAfterSeconds,
            headers: { "retry-after": "9".repeat(400) },
            name: "delta seconds too large to represent",
        },
        {
            expected: MaximumRetryAfterSeconds,
            headers: { "retry-after": "Sat, 01 Jan 2028 00:00:00 GMT" },
            name: "http date beyond the maximum",
        },
        { expected: DefaultRetryAfterSeconds, headers: {}, name: "missing header" },
        { expected: DefaultRetryAfterSeconds, headers: { "retry-after": "" }, name: "empty header" },
        { expected: DefaultRetryAfterSeconds, headers: { "retry-after": "invalid" }, name: "invalid header" },
    ] as { expected: number; headers: Record<string, string>; name: string }[])(
        "returns the retry after for $name",
        ({ expected, headers }) => {
            expect(getRetryAfter(newResponse(429, headers))).toBe(expected);
        },
    );

    it("rounds a partial second up", () => {
        vi.setSystemTime(new Date("2026-01-01T00:00:59.500Z"));

        expect(getRetryAfter(newResponse(429, { "retry-after": "Thu, 01 Jan 2026 00:01:00 GMT" }))).toBe(1);
    });
});

describe("registerRateLimitInterceptor", () => {
    it("throws a RateLimitedError when the response is rejected with a 429", async () => {
        const instance = newInstance(429, { "retry-after": "30" });

        const err = await instance.get("/api").catch((e: unknown) => e);

        expect(err).toBeInstanceOf(RateLimitedError);
        expect((err as RateLimitedError).retryAfter).toBe(30);
        expect((err as RateLimitedError).message).toBe("Too Many Requests");
    });

    it("throws a RateLimitedError when validateStatus accepts a 429", async () => {
        const instance = newInstance(429, { "retry-after": "15" }, () => true);

        const err = await instance.get("/api").catch((e: unknown) => e);

        expect(err).toBeInstanceOf(RateLimitedError);
        expect((err as RateLimitedError).retryAfter).toBe(15);
    });

    it("uses the default retry after when the header is missing", async () => {
        const instance = newInstance(429);

        const err = await instance.get("/api").catch((e: unknown) => e);

        expect(err).toBeInstanceOf(RateLimitedError);
        expect((err as RateLimitedError).retryAfter).toBe(DefaultRetryAfterSeconds);
    });

    it("passes through other errors unchanged", async () => {
        const instance = newInstance(401);

        const err = await instance.get("/api").catch((e: unknown) => e);

        expect(err).toBeInstanceOf(AxiosError);
        expect(err).not.toBeInstanceOf(RateLimitedError);
        expect((err as AxiosError).response?.status).toBe(401);
    });

    it("passes through successful responses unchanged", async () => {
        const instance = newInstance(200);

        const resp = await instance.get("/api");

        expect(resp.status).toBe(200);
    });
});
