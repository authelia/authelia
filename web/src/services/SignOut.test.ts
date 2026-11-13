// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { DeleteWithOptionalResponse, Get, PostWithOptionalResponse } from "@services/Client";
import { cancelSignOut, getSignOutPending, signOut } from "@services/SignOut";

vi.mock("@services/Api", () => ({
    LogoutPath: "/logout",
}));
vi.mock("@services/Client", () => ({
    DeleteWithOptionalResponse: vi.fn(),
    Get: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
}));

it("signs out with target URL", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue({ safeTargetURL: true });
    const result = await signOut("https://example.com");
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/logout", { targetURL: "https://example.com" }, undefined);
    expect(result).toEqual({ safeTargetURL: true });
});

it("signs out without target URL", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    const result = await signOut(undefined);
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/logout", {}, undefined);
    expect(result).toBeUndefined();
});

it("forwards the abort signal when provided", async () => {
    const signal = new AbortController().signal;
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    await signOut("https://example.com", signal);
    expect(PostWithOptionalResponse).toHaveBeenLastCalledWith("/logout", { targetURL: "https://example.com" }, signal);
});

it("signs out with the flow id of the logout being confirmed", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue({
        redirectURL: "https://app.example.com",
        safeTargetURL: false,
    });
    await signOut(undefined, undefined, "flow-1");
    expect(PostWithOptionalResponse).toHaveBeenLastCalledWith("/logout", { flowID: "flow-1" }, undefined);
});

it("reports the pending logout of a flow", async () => {
    (Get as any).mockResolvedValue({ pending: true });
    const result = await getSignOutPending("flow 1&x");
    expect(Get).toHaveBeenLastCalledWith("/logout?flow_id=flow+1%26x", undefined);
    expect(result).toEqual({ pending: true });
});

it("cancels the pending logout of a flow", async () => {
    (DeleteWithOptionalResponse as any).mockResolvedValue(undefined);
    await cancelSignOut("flow-1");
    expect(DeleteWithOptionalResponse).toHaveBeenLastCalledWith("/logout?flow_id=flow-1", undefined, undefined);
});
