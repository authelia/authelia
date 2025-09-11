import axios from "axios";
import { vi } from "vitest";

import { ScopeDescription } from "@components/OpenIDConnect";
import { Get, Post } from "@services/Client";
import * as Consent from "@services/ConsentOpenIDConnect";

vi.mock("axios");
vi.mock("@components/OpenIDConnect", () => ({
    ScopeDescription: vi.fn(),
}));
vi.mock("@constants/SearchParams", () => ({
    FlowID: "flow_id",
    UserCode: "user_code",
}));
vi.mock("@services/Api", () => ({
    OpenIDConnectConsentPath: "/consent",
    OpenIDConnectDeviceAuthorizationPath: "/device",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
    Post: vi.fn(),
}));

it("gets consent response without params", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await Consent.getConsentResponse();
    expect(Get).toHaveBeenCalledWith("/consent?");
    expect(result).toBe("response");
});

it("gets consent response with flow id", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await Consent.getConsentResponse("flow123");
    expect(Get).toHaveBeenCalledWith("/consent?flow_id=flow123");
    expect(result).toBe("response");
});

it("gets consent response with user code", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await Consent.getConsentResponse(undefined, "code123");
    expect(Get).toHaveBeenCalledWith("/consent?user_code=code123");
    expect(result).toBe("response");
});

it("gets consent response with both params", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await Consent.getConsentResponse("flow123", "code123");
    expect(Get).toHaveBeenCalledWith("/consent?flow_id=flow123&user_code=code123");
    expect(result).toBe("response");
});

it("posts consent response accept", async () => {
    (Post as any).mockResolvedValue("response");
    const result = await Consent.postConsentResponseAccept(true, "client", ["claim1"], "flow", "sub", "code");
    expect(Post).toHaveBeenCalledWith("/consent", {
        flow_id: "flow",
        client_id: "client",
        consent: true,
        pre_configure: true,
        claims: ["claim1"],
        subflow: "sub",
        user_code: "code",
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
    const result = await Consent.postConsentResponseReject("client", "flow", "sub", "code");
    expect(Post).toHaveBeenCalledWith("/consent", {
        flow_id: "flow",
        client_id: "client",
        consent: false,
        pre_configure: false,
        subflow: "sub",
        user_code: "code",
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
    expect(Consent.getClaimDescription("other_claim")).toBe("Other claim");
});
