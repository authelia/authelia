// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import axios from "axios";

import { ScopeDescription } from "@components/OpenIDConnect";
import { Get, Post } from "@services/Client";
import * as Consent from "@services/ConsentOpenIDConnect";

vi.mock("axios");
vi.mock("@components/OpenIDConnect", () => ({
    ScopeDescription: vi.fn(),
}));
vi.mock("@constants/SearchParams", () => ({
    FlowID: "flow_id",
    SubFlow: "subflow",
    UserCode: "user_code",
}));
vi.mock("@services/Api", () => ({
    OpenIDConnectConsentPath: "/consent",
    OpenIDConnectConsentUserCodePath: "/consent/user-code",
    OpenIDConnectDeviceAuthorizationPath: "/device",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
    Post: vi.fn(),
}));

it("gets consent response with flow id", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await Consent.getConsentResponse("flow123");
    expect(Get).toHaveBeenCalledWith("/consent?flow_id=flow123");
    expect(result).toBe("response");
});

it("gets consent response with flow id and subflow", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await Consent.getConsentResponse("flow123", "device_authorization");
    expect(Get).toHaveBeenCalledWith("/consent?flow_id=flow123&subflow=device_authorization");
    expect(result).toBe("response");
});

it("posts the user code", async () => {
    (Post as any).mockResolvedValue("response");
    const result = await Consent.postConsentUserCode("code123");
    expect(Post).toHaveBeenCalledWith("/consent/user-code", { user_code: "code123" });
    expect(result).toBe("response");
});

it("posts consent response accept", async () => {
    (Post as any).mockResolvedValue("response");
    const result = await Consent.postConsentResponseAccept(true, "client", ["claim1"], "flow", "sub");
    expect(Post).toHaveBeenCalledWith("/consent", {
        claims: ["claim1"],
        client_id: "client",
        consent: true,
        flow_id: "flow",
        pre_configure: true,
        subflow: "sub",
    });
    expect(result).toBe("response");
});

it("puts device code flow user code", async () => {
    (axios.put as any).mockResolvedValue("response");
    const result = await Consent.putDeviceCodeFlowUserCode("flow", "code");
    expect(axios.put).toHaveBeenCalledWith("/device", expect.any(URLSearchParams));
    expect(result).toBe("response");
});

it("posts consent response reject", async () => {
    (Post as any).mockResolvedValue("response");
    const result = await Consent.postConsentResponseReject("client", "flow", "sub");
    expect(Post).toHaveBeenCalledWith("/consent", {
        client_id: "client",
        consent: false,
        flow_id: "flow",
        pre_configure: false,
        subflow: "sub",
    });
    expect(result).toBe("response");
});

it("formats scope when not starting with scopes.", () => {
    expect(Consent.formatScope("openid", "fallback")).toBe("openid");
});

it("formats scope when empty", () => {
    (ScopeDescription as any).mockReturnValue("fallback desc");
    expect(Consent.formatScope("", "fallback")).toBe("fallback desc");
});

it("formats scope when starting with scopes.", () => {
    (ScopeDescription as any).mockReturnValue("fallback desc");
    expect(Consent.formatScope("scopes.openid", "fallback")).toBe("fallback desc");
});

it("formats claim when not starting with claims.", () => {
    expect(Consent.formatClaim("name", "fallback")).toBe("name");
});

it("formats claim when starting with claims.", () => {
    expect(Consent.formatClaim("claims.name", "fallback")).toBe("Fallback");
});

it("gets claim description for name", () => {
    expect(Consent.getClaimDescription("name")).toBe("Display Name");
});

it("gets claim description for sub", () => {
    expect(Consent.getClaimDescription("sub")).toBe("Unique Identifier");
});

it("gets claim description for zoneinfo", () => {
    expect(Consent.getClaimDescription("zoneinfo")).toBe("Timezone");
});

it("gets claim description for locale", () => {
    expect(Consent.getClaimDescription("locale")).toBe("Locale / Language");
});

it("gets claim description for updated_at", () => {
    expect(Consent.getClaimDescription("updated_at")).toBe("Information Updated Time");
});

it("gets claim description for profile", () => {
    expect(Consent.getClaimDescription("profile")).toBe("Profile URL");
});

it("gets claim description for website", () => {
    expect(Consent.getClaimDescription("website")).toBe("Website URL");
});

it("gets claim description for picture", () => {
    expect(Consent.getClaimDescription("picture")).toBe("Picture URL");
});

it("gets claim description for other", () => {
    expect(Consent.getClaimDescription("other_claim")).toBe("Other Claim");
});
