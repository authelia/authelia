// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { AxiosResponse } from "axios";

import {
    hasServiceError,
    toData,
    validateStatusAuthentication,
    validateStatusOneTimeCode,
    validateStatusWebAuthnCreation,
} from "@services/Api";

it("reports no error for ok response", () => {
    const resp = { data: { data: "test", status: "OK" }, status: 200 } as AxiosResponse;
    expect(hasServiceError(resp)).toEqual({ errored: false, message: null });
});

it("reports error for ko response", () => {
    const resp = { data: { message: "bad request", status: "KO" }, status: 400 } as AxiosResponse;
    expect(hasServiceError(resp)).toEqual({ errored: true, message: "bad request" });
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

describe("toData", () => {
    it("returns the payload for a successful response", () => {
        const resp = { data: { data: { value: 1 }, status: "OK" }, headers: {}, status: 200 } as any;
        expect(toData(resp)).toEqual({ value: 1 });
    });

    it("returns undefined for a failed response", () => {
        const resp = { data: { status: "KO" }, headers: {}, status: 400 } as any;
        expect(toData(resp)).toBeUndefined();
    });

    it("returns undefined without a status", () => {
        const resp = { data: { value: 1 }, headers: {}, status: 200 } as any;
        expect(toData(resp)).toBeUndefined();
    });

    it("returns undefined without data", () => {
        const resp = { data: undefined, headers: {}, status: 200 } as any;
        expect(toData(resp)).toBeUndefined();
    });
});
