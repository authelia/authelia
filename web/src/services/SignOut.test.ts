// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { PostWithOptionalResponse } from "@services/Client";
import { signOut } from "@services/SignOut";

vi.mock("@services/Api", () => ({
    LogoutPath: "/logout",
}));
vi.mock("@services/Client", () => ({
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

it("signs out with flow id", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    await signOut(undefined, undefined, "6c2b5c4e-1f0a-4f6e-9a5c-3d7b2e1f0a9c");
    expect(PostWithOptionalResponse).toHaveBeenLastCalledWith(
        "/logout",
        { flowID: "6c2b5c4e-1f0a-4f6e-9a5c-3d7b2e1f0a9c" },
        undefined,
    );
});

it("forwards the abort signal when provided", async () => {
    const signal = new AbortController().signal;
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    await signOut("https://example.com", signal);
    expect(PostWithOptionalResponse).toHaveBeenLastCalledWith("/logout", { targetURL: "https://example.com" }, signal);
});
