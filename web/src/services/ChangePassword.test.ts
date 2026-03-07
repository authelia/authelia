import { postPasswordChange } from "@services/ChangePassword";
import { PostWithOptionalResponse } from "@services/Client";

vi.mock("@services/Api", () => ({
    ChangePasswordPath: "/change-password",
}));

vi.mock("@services/Client", () => ({
    PostWithOptionalResponse: vi.fn(),
}));

it("calls PostWithOptionalResponse with correct data", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue("success");

    const result = await postPasswordChange("user", "old", "new");

    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/change-password", {
        new_password: "new",
        old_password: "old",
        username: "user",
    });
    expect(result).toBe("success");
});
