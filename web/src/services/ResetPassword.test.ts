import axios from "axios";
import { vi } from "vitest";

import { PostWithOptionalResponse, PostWithOptionalResponseRateLimited } from "@services/Client";
import {
    completeResetPasswordProcess,
    deleteResetPasswordToken,
    initiateResetPasswordProcess,
    resetPassword,
} from "@services/ResetPassword";

vi.mock("axios");
vi.mock("@services/Api", () => ({
    CompleteResetPasswordPath: "/reset-password/finish",
    InitiateResetPasswordPath: "/reset-password/start",
    ResetPasswordPath: "/reset-password",
    validateStatusTooManyRequests: vi.fn(),
}));
vi.mock("@services/Client", () => ({
    PostWithOptionalResponse: vi.fn(),
    PostWithOptionalResponseRateLimited: vi.fn(),
}));

it("initiates reset password process", async () => {
    (PostWithOptionalResponseRateLimited as any).mockResolvedValue("response");
    const result = await initiateResetPasswordProcess("user");
    expect(PostWithOptionalResponseRateLimited).toHaveBeenCalledWith("/reset-password/start", { username: "user" });
    expect(result).toBe("response");
});

it("completes reset password process", async () => {
    (PostWithOptionalResponseRateLimited as any).mockResolvedValue("response");
    const result = await completeResetPasswordProcess("token123");
    expect(PostWithOptionalResponseRateLimited).toHaveBeenCalledWith("/reset-password/finish", { token: "token123" });
    expect(result).toBe("response");
});

it("resets password", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue("response");
    const result = await resetPassword("newpass");
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/reset-password", { password: "newpass" });
    expect(result).toBe("response");
});

it("deletes reset password token with ok", async () => {
    const mockRes = { status: 200, data: { status: "OK" } };
    (axios as any).mockResolvedValue(mockRes);
    const result = await deleteResetPasswordToken("token123");
    expect(axios).toHaveBeenCalledWith({
        method: "DELETE",
        url: "/reset-password",
        data: { token: "token123" },
        validateStatus: expect.any(Function),
    });
    expect(result).toEqual({ ok: true, status: 200 });
});

it("deletes reset password token with error", async () => {
    const mockRes = { status: 400, data: { status: "KO" } };
    (axios as any).mockResolvedValue(mockRes);
    const result = await deleteResetPasswordToken("token123");
    expect(result).toEqual({ ok: false, status: 400 });
});
