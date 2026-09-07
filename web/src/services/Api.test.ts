import { AxiosResponse } from "axios";

import {
    hasServiceError,
    toDataRateLimited,
    validateStatusAuthentication,
    validateStatusOneTimeCode,
    validateStatusTooManyRequests,
    validateStatusWebAuthnCreation,
} from "@services/Api";

it("throws for missing retry-after header", () => {
    const resp = { data: { status: "KO" }, headers: {}, status: 429 } as any;
    expect(() => toDataRateLimited(resp)).toThrow("Header Retry-After is missing");
});

it("handles numeric retry-after", () => {
    const resp = { data: { status: "KO" }, headers: { "retry-after": "30" }, status: 429 } as any;
    const result = toDataRateLimited(resp);
    expect(result?.limited).toBe(true);
    expect(result?.retryAfter).toBe(30);
});

it("handles date retry-after", () => {
    const future = new Date(Date.now() + 60000).toUTCString();
    const resp = { data: { status: "KO" }, headers: { "retry-after": future }, status: 429 } as any;
    const result = toDataRateLimited(resp);
    expect(result?.limited).toBe(true);
    expect(result?.retryAfter).toBeGreaterThan(50);
});

it("throws for invalid date retry-after", () => {
    const resp = { data: { status: "KO" }, headers: { "retry-after": "invalid-date" }, status: 429 } as any;
    expect(() => toDataRateLimited(resp)).toThrow("Header Retry-After has an invalid date value");
});

it("returns data for ok status", () => {
    const resp = { data: { data: "test", status: "OK" }, status: 200 } as AxiosResponse;
    expect(toDataRateLimited(resp)).toEqual({ data: "test", limited: false, retryAfter: 0 });
});

it("returns limited for ko status with 429", () => {
    const resp = { data: { status: "KO" }, headers: { "retry-after": "30" }, status: 429 } as any;
    expect(toDataRateLimited(resp)).toEqual({ limited: true, retryAfter: 30 });
});

it("returns limited for 429 status", () => {
    const resp = { data: { status: "OTHER" }, headers: { "retry-after": "30" }, status: 429 } as any;
    expect(toDataRateLimited(resp)).toEqual({ limited: true, retryAfter: 30 });
});

it("returns undefined for no data", () => {
    const resp = { data: null, status: 200 } as AxiosResponse;
    expect(toDataRateLimited(resp)).toBeUndefined();
});

it("reports no error for ok response", () => {
    const resp = { data: { data: "test", status: "OK" }, status: 200 } as AxiosResponse;
    expect(hasServiceError(resp)).toEqual({ errored: false, message: null });
});

it("reports error for ko response", () => {
    const resp = { data: { message: "bad request", status: "KO" }, status: 400 } as AxiosResponse;
    expect(hasServiceError(resp)).toEqual({ errored: true, message: "bad request" });
});

it("validates status for too many requests", () => {
    expect(validateStatusTooManyRequests(200)).toBe(true);
    expect(validateStatusTooManyRequests(299)).toBe(true);
    expect(validateStatusTooManyRequests(429)).toBe(true);
    expect(validateStatusTooManyRequests(400)).toBe(false);
    expect(validateStatusTooManyRequests(500)).toBe(false);
});

it("validates status for authentication", () => {
    expect(validateStatusAuthentication(200)).toBe(true);
    expect(validateStatusAuthentication(401)).toBe(true);
    expect(validateStatusAuthentication(403)).toBe(true);
    expect(validateStatusAuthentication(400)).toBe(false);
    expect(validateStatusAuthentication(500)).toBe(false);
});

it("validates status for one time code", () => {
    expect(validateStatusOneTimeCode(200)).toBe(true);
    expect(validateStatusOneTimeCode(401)).toBe(true);
    expect(validateStatusOneTimeCode(403)).toBe(true);
    expect(validateStatusOneTimeCode(399)).toBe(true);
    expect(validateStatusOneTimeCode(400)).toBe(false);
    expect(validateStatusOneTimeCode(500)).toBe(false);
});

it("validates status for webauthn creation", () => {
    expect(validateStatusWebAuthnCreation(200)).toBe(true);
    expect(validateStatusWebAuthnCreation(299)).toBe(true);
    expect(validateStatusWebAuthnCreation(409)).toBe(true);
    expect(validateStatusWebAuthnCreation(400)).toBe(false);
    expect(validateStatusWebAuthnCreation(500)).toBe(false);
});
