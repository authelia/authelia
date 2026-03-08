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
    expect(PostWithOptionalResponseRateLimited).toHaveBeenCalledWith("/duo/signin", {
        flow: "flowtype",
        flowID: "flow",
        subflow: "sub",
        targetURL: "url",
        userCode: "code",
    });
    expect(result).toBe("response");
});

it("initiates duo device selection process", async () => {
    (Get as any).mockResolvedValue("response");
    const result = await initiateDuoDeviceSelectionProcess();
    expect(Get).toHaveBeenCalledWith("/duo/devices");
    expect(result).toBe("response");
});

it("completes duo device selection process", async () => {
    (PostWithOptionalResponse as any).mockResolvedValue("response");
    const result = await completeDuoDeviceSelectionProcess({ device: "dev", method: "push" });
    expect(PostWithOptionalResponse).toHaveBeenCalledWith("/duo/device", {
        device: "dev",
        method: "push",
    });
    expect(result).toBe("response");
});

it("gets preferred duo device", async () => {
    const { getPreferredDuoDevice } = await import("@services/PushNotification");
    (Get as any).mockResolvedValue("response");
    const result = await getPreferredDuoDevice();
    expect(Get).toHaveBeenCalledWith("/duo/signin");
    expect(result).toBe("response");
});
