// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { ErrorCodeTranslationKeys, isErrorCode, translateErrorCode } from "@services/ErrorCode";

import portal from "../../../internal/server/locales/en/portal.json";

const translate = vi.fn((key: string) => `t:${key}`);

beforeEach(() => {
    translate.mockClear();
});

it("translates a known code in the portal namespace", () => {
    expect(translateErrorCode(translate, "password_policy", "fallback")).toBe(
        "t:Your supplied password does not meet the password policy requirements",
    );
    expect(translate).toHaveBeenCalledWith("Your supplied password does not meet the password policy requirements", {
        ns: "portal",
    });
});

it("returns the fallback for an unknown code", () => {
    expect(translateErrorCode(translate, "not_a_code", "fallback")).toBe("fallback");
    expect(translate).not.toHaveBeenCalled();
});

it("returns the fallback for a missing code", () => {
    expect(translateErrorCode(translate, undefined, "fallback")).toBe("fallback");
});

it("does not treat prototype properties as codes", () => {
    expect(isErrorCode("toString")).toBe(false);
});

it("has an english portal locale entry for every code", () => {
    const keys: Record<string, string> = portal;

    for (const key of Object.values(ErrorCodeTranslationKeys)) {
        expect(keys).toHaveProperty([key]);
    }
});
