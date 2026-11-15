// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { Get, Post, PostWithOptionalResponse } from "@services/Client";
import { generateRecoveryCodes, getRecoveryCodesStatus, signInWithRecoveryCode } from "@services/UserRecoveryCodes";

vi.mock("@services/Api", () => ({
    CompleteRecoveryCodeSignInPath: "/recovery-code",
    RecoveryCodesGeneratePath: "/recovery-codes/generate",
    RecoveryCodesStatusPath: "/recovery-codes",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
    Post: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
}));

beforeEach(() => {
    vi.clearAllMocks();
});

it("maps the status payload timestamps to dates", async () => {
    vi.mocked(Get).mockResolvedValue({
        codes_remaining: 7,
        codes_total: 10,
        generated_at: "2026-01-02T03:04:05Z",
        last_used_at: "2026-02-03T04:05:06Z",
    });

    const result = await getRecoveryCodesStatus();

    expect(Get).toHaveBeenCalledWith("/recovery-codes");
    expect(result).toEqual({
        codes_remaining: 7,
        codes_total: 10,
        generated_at: new Date("2026-01-02T03:04:05Z"),
        last_used_at: new Date("2026-02-03T04:05:06Z"),
    });
    expect(result.generated_at).toBeInstanceOf(Date);
    expect(result.last_used_at).toBeInstanceOf(Date);
});

it("leaves missing status timestamps undefined", async () => {
    vi.mocked(Get).mockResolvedValue({ codes_remaining: 0, codes_total: 0 });

    const result = await getRecoveryCodesStatus();

    expect(result).toEqual({
        codes_remaining: 0,
        codes_total: 0,
        generated_at: undefined,
        last_used_at: undefined,
    });
});

it("propagates status retrieval errors", async () => {
    vi.mocked(Get).mockRejectedValue(new Error("boom"));

    await expect(getRecoveryCodesStatus()).rejects.toThrow("boom");
});

it("generates recovery codes", async () => {
    const response = { codes: ["AAAA-BBBB"], notification_sent: true };
    vi.mocked(Post).mockResolvedValue(response);

    const result = await generateRecoveryCodes();

    expect(Post).toHaveBeenCalledWith("/recovery-codes/generate");
    expect(result).toBe(response);
});

it("signs in with a recovery code", async () => {
    vi.mocked(PostWithOptionalResponse).mockResolvedValue({ redirect: "https://example.com" });

    const result = await signInWithRecoveryCode({
        code: "AAAA-BBBB",
        flow: "openid_connect",
        flowID: "flow-id",
        subflow: "subflow",
        targetURL: "https://target.example.com",
        userCode: "user-code",
    });

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/recovery-code", {
        code: "AAAA-BBBB",
        flow: "openid_connect",
        flowID: "flow-id",
        subflow: "subflow",
        targetURL: "https://target.example.com",
        userCode: "user-code",
    });
    expect(result).toEqual({ redirect: "https://example.com" });
});

it("signs in with only a recovery code", async () => {
    vi.mocked(PostWithOptionalResponse).mockResolvedValue(undefined);

    const result = await signInWithRecoveryCode({ code: "AAAA-BBBB" });

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/recovery-code", {
        code: "AAAA-BBBB",
        flow: undefined,
        flowID: undefined,
        subflow: undefined,
        targetURL: undefined,
        userCode: undefined,
    });
    expect(result).toBeUndefined();
});
