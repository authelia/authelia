import { vi } from "vitest";

import { PostWithOptionalResponse } from "@services/Client";
import { postFirstFactor, postFirstFactorReauthenticate, postSecondFactor } from "@services/Password";

vi.mock("@services/Api", () => ({
    CompletePasswordSignInPath: "/password/signin",
    FirstFactorPath: "/firstfactor",
    FirstFactorReauthenticatePath: "/firstfactor/reauth",
}));
vi.mock("@services/Client", () => ({
    PostWithOptionalResponse: vi.fn(),
}));

it("posts first factor with response", async () => {
    const mockResponse = { token: "abc" };
    (PostWithOptionalResponse as any).mockResolvedValue(mockResponse);
    const result = await postFirstFactor("user", "pass", true, "url", "POST", "flow", "flowtype", "sub", "code");
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/firstfactor", {
        username: "user",
        password: "pass",
        keepMeLoggedIn: true,
        targetURL: "url",
        requestMethod: "POST",
        flowID: "flow",
        flow: "flowtype",
        subflow: "sub",
        userCode: "code",
    });
    expect(result).toEqual(mockResponse);
});

it("posts first factor without response", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    const result = await postFirstFactor("user", "pass", false);
    expect(result).toEqual({});
});

it("posts first factor reauthenticate with response", async () => {
    const mockResponse = { token: "abc" };
    (PostWithOptionalResponse as any).mockResolvedValue(mockResponse);
    const result = await postFirstFactorReauthenticate("pass", "url", "POST", "flow", "flowtype", "sub", "code");
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/firstfactor/reauth", {
        password: "pass",
        targetURL: "url",
        requestMethod: "POST",
        flowID: "flow",
        flow: "flowtype",
        subflow: "sub",
        userCode: "code",
    });
    expect(result).toEqual(mockResponse);
});

it("posts first factor reauthenticate without response", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    const result = await postFirstFactorReauthenticate("pass");
    expect(result).toEqual({});
});

it("posts second factor with response", async () => {
    const mockResponse = { token: "abc" };
    (PostWithOptionalResponse as any).mockResolvedValue(mockResponse);
    const result = await postSecondFactor("pass", "url", "flow", "flowtype", "sub");
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/password/signin", {
        password: "pass",
        targetURL: "url",
        flowID: "flow",
        flow: "flowtype",
        subflow: "sub",
    });
    expect(result).toEqual(mockResponse);
});

it("posts second factor without response", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue(undefined);
    const result = await postSecondFactor("pass");
    expect(result).toEqual({});
});
