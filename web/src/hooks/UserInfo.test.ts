import { renderHook } from "@testing-library/react";
import { vi } from "vitest";

import { useRemoteCall } from "@hooks/RemoteCall";
import { useUserInfoGET, useUserInfoPOST } from "@hooks/UserInfo";
import { getUserInfo, postUserInfo } from "@services/UserInfo";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("user info post call", () => {
    (useRemoteCall as any).mockReturnValue("postResult");
    const { result } = renderHook(() => useUserInfoPOST());
    expect(useRemoteCall).toHaveBeenCalledWith(postUserInfo, []);
    expect(result.current).toBe("postResult");
});

it("user info get call", () => {
    (useRemoteCall as any).mockReturnValue("getResult");
    const { result } = renderHook(() => useUserInfoGET());
    expect(useRemoteCall).toHaveBeenCalledWith(getUserInfo, []);
    expect(result.current).toBe("getResult");
});
