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
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/logout", { targetURL: "https://example.com" });
    expect(result).toEqual({ safeTargetURL: true });
});

it("signs out without target URL", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    const result = await signOut(undefined);
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/logout", {});
    expect(result).toBeUndefined();
});
