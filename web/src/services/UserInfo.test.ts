import { vi } from "vitest";

import { SecondFactorMethod } from "@models/Methods";
import { Get, Post, PostWithOptionalResponse } from "@services/Client";
import {
    getUserInfo,
    isMethod2FA,
    postUserInfo,
    setPreferred2FAMethod,
    toMethod2FA,
    toSecondFactorMethod,
} from "@services/UserInfo";

vi.mock("@models/Methods", () => ({
    SecondFactorMethod: {
        TOTP: "totp",
        WebAuthn: "webauthn",
        MobilePush: "mobile_push",
    },
}));
vi.mock("@services/Api", () => ({
    UserInfo2FAMethodPath: "/user/2fa",
    UserInfoPath: "/user/info",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
    Post: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
}));

it("checks if method is 2fa", () => {
    expect(isMethod2FA("webauthn")).toBe(true);
    expect(isMethod2FA("totp")).toBe(true);
    expect(isMethod2FA("mobile_push")).toBe(true);
    expect(isMethod2FA("unknown")).toBe(false);
});

it("converts to second factor method", () => {
    expect(toSecondFactorMethod("totp")).toBe(SecondFactorMethod.TOTP);
    expect(toSecondFactorMethod("webauthn")).toBe(SecondFactorMethod.WebAuthn);
    expect(toSecondFactorMethod("mobile_push")).toBe(SecondFactorMethod.MobilePush);
});

it("converts to method 2fa", () => {
    expect(toMethod2FA(SecondFactorMethod.TOTP)).toBe("totp");
    expect(toMethod2FA(SecondFactorMethod.WebAuthn)).toBe("webauthn");
    expect(toMethod2FA(SecondFactorMethod.MobilePush)).toBe("mobile_push");
});

it("posts user info", async () => {
    const mockRes = { display_name: "user", emails: ["a@b.com"], method: "totp" as const };
    (Post as any).mockResolvedValue(mockRes);
    const result = await postUserInfo();
    expect(Post).toHaveBeenCalledWith("/user/info");
    expect(result).toEqual({ ...mockRes, method: SecondFactorMethod.TOTP });
});

it("gets user info", async () => {
    const mockRes = { display_name: "user", emails: ["a@b.com"], method: "webauthn" as const };
    (Get as any).mockResolvedValue(mockRes);
    const result = await getUserInfo();
    expect(Get).toHaveBeenCalledWith("/user/info");
    expect(result).toEqual({ ...mockRes, method: SecondFactorMethod.WebAuthn });
});

it("sets preferred 2fa method", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue("response");
    const result = await setPreferred2FAMethod(SecondFactorMethod.MobilePush);
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/user/2fa", { method: "mobile_push" });
    expect(result).toBe("response");
});
