import { AxiosResponse } from "axios";

import { toDataRateLimited } from "@services/Api";

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
