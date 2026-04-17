import { Get, PostWithOptionalResponse, PostWithOptionalResponseRateLimited } from "@services/Client";
import {
    completeDuoDeviceSelectionProcess,
    completePushNotificationSignIn,
    initiateDuoDeviceSelectionProcess,
} from "@services/PushNotification";

vi.mock("@services/Api", () => ({
    CompleteDuoDeviceSelectionPath: "/duo/device",
    CompletePushNotificationSignInPath: "/duo/signin",
    InitiateDuoDeviceSelectionPath: "/duo/devices",
}));
vi.mock("@services/Client", () => ({
    Get: vi.fn(),
    PostWithOptionalResponse: vi.fn(),
    PostWithOptionalResponseRateLimited: vi.fn(),
}));

it("completes push notification sign in", async () => {
    (PostWithOptionalResponseRateLimited as any).mockResolvedValue("response");
    const result = await completePushNotificationSignIn("url", "flow", "flowtype", "sub", "code");
    expect(PostWithOptionalResponseRateLimited).toHaveBeenCalledWith(
        "/duo/signin",
        {
            flow: "flowtype",
            flowID: "flow",
            subflow: "sub",
            targetURL: "url",
            userCode: "code",
        },
        undefined,
    );
    expect(result).toBe("response");
});

it("initiates duo device selection process", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await initiateDuoDeviceSelectionProcess();
    expect(Get).toHaveBeenCalledWith("/duo/devices", undefined);
    expect(result).toBe("response");
});

it("completes duo device selection process", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue("response");
    const result = await completeDuoDeviceSelectionProcess({ device: "dev", method: "push" });
    expect(PostWithOptionalResponse).toHaveBeenCalledWith(
        "/duo/device",
        {
            device: "dev",
            method: "push",
        },
        undefined,
    );
    expect(result).toBe("response");
});

it("gets preferred duo device", async () => {
    const { getPreferredDuoDevice } = await import("@services/PushNotification");
    (Get as any).mockResolvedValue("response");
    const result = await getPreferredDuoDevice();
    expect(Get).toHaveBeenCalledWith("/duo/signin", undefined);
    expect(result).toBe("response");
});

it("forwards the abort signal when provided", async () => {
    const { getPreferredDuoDevice } = await import("@services/PushNotification");
    const signal = new AbortController().signal;
    (PostWithOptionalResponseRateLimited as any).mockResolvedValue("response");
    (PostWithOptionalResponse as any).mockResolvedValue("response");
    (Get as any).mockResolvedValue("response");

    await completePushNotificationSignIn("url", "flow", "flowtype", "sub", "code", signal);
    expect(PostWithOptionalResponseRateLimited).toHaveBeenLastCalledWith("/duo/signin", expect.any(Object), signal);

    await initiateDuoDeviceSelectionProcess(signal);
    expect(Get).toHaveBeenLastCalledWith("/duo/devices", signal);

    await completeDuoDeviceSelectionProcess({ device: "dev", method: "push" }, signal);
    expect(PostWithOptionalResponse).toHaveBeenLastCalledWith("/duo/device", expect.any(Object), signal);

    await getPreferredDuoDevice(signal);
    expect(Get).toHaveBeenLastCalledWith("/duo/signin", signal);
});
