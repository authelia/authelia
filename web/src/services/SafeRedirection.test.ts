import { PostWithOptionalResponse } from "@services/Client";
import { checkSafeRedirection } from "@services/SafeRedirection";

vi.mock("@services/Api", () => ({
    ChecksSafeRedirectionPath: "/checks/safe-redirection",
}));
vi.mock("@services/Client", () => ({
    PostWithOptionalResponse: vi.fn(),
}));

it("posts uri for safe redirection check", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue({ ok: true });
    const result = await checkSafeRedirection("https://example.com");
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/checks/safe-redirection", {
        uri: "https://example.com",
    });
    expect(result).toEqual({ ok: true });
});
