import { vi } from "vitest";

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
        username: "user",
        old_password: "old",
        new_password: "new",
    });
    expect(result).toBe("success");
});
