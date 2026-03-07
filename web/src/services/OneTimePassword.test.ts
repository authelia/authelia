import {
    DeleteWithOptionalResponse,
    PostWithOptionalResponse,
    PostWithOptionalResponseRateLimited,
} from "@services/Client";
import { completeTOTPRegister, completeTOTPSignIn, stopTOTPRegister } from "@services/OneTimePassword";

vi.mock("@services/Api", () => ({
    CompleteTOTPSignInPath: "/totp/signin",
    TOTPRegistrationPath: "/totp/register",
}));
vi.mock("@services/Client", () => ({
    DeleteWithOptionalResponse: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
    PostWithOptionalResponseRateLimited: vi.fn(),
}));

it("completes totp sign in", async () => {
    (PostWithOptionalResponseRateLimited as any).mockResolvedValue("response");
    const result = await completeTOTPSignIn("123456", "url", "flow", "flowtype", "sub", "code");
    expect(PostWithOptionalResponseRateLimited).toHaveBeenCalledWith("/totp/signin", {
        flow: "flowtype",
        flowID: "flow",
        subflow: "sub",
        targetURL: "url",
        token: "123456",
        userCode: "code",
    });
    expect(result).toBe("response");
});

it("completes totp register", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue("response");
    const result = await completeTOTPRegister("123456");
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/totp/register", {
        token: "123456",
    });
    expect(result).toBe("response");
});

it("stops totp register", async () => {
    (DeleteWithOptionalResponse as any).mockResolvedValue("response");
    const result = await stopTOTPRegister();
    expect(DeleteWithOptionalResponse).toHaveBeenCalledWith("/totp/register");
    expect(result).toBe("response");
});
