import { renderHook } from "@testing-library/react";

import { useRemoteCall } from "@hooks/RemoteCall";
import {
    useAdminConfigurationGET,
    useAllUserInfoGET,
    useUserGET,
    useUserManagementAttributeMetadataGET,
} from "@hooks/UserManagement";
import { getAdminConfiguration, getAllUserInfo, getUser, getUserAttributeMetadata } from "@services/UserManagement";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

vi.mock("@services/UserManagement", () => ({
    getAdminConfiguration: vi.fn(),
    getAllUserInfo: vi.fn(),
    getUser: vi.fn(),
    getUserAttributeMetadata: vi.fn(),
}));

beforeEach(() => {
    vi.clearAllMocks();
    vi.mocked(useRemoteCall).mockReturnValue("result" as any);
});

it("calls useRemoteCall with getAllUserInfo", () => {
    const { result } = renderHook(() => useAllUserInfoGET());

    expect(useRemoteCall).toHaveBeenCalledWith(getAllUserInfo);
    expect(result.current).toBe("result");
});

it("calls useRemoteCall with getAdminConfiguration", () => {
    const { result } = renderHook(() => useAdminConfigurationGET());

    expect(useRemoteCall).toHaveBeenCalledWith(getAdminConfiguration);
    expect(result.current).toBe("result");
});

it("calls useRemoteCall with getUserAttributeMetadata", () => {
    const { result } = renderHook(() => useUserManagementAttributeMetadataGET());

    expect(useRemoteCall).toHaveBeenCalledWith(getUserAttributeMetadata);
    expect(result.current).toBe("result");
});

it("calls useRemoteCall with a getUser closure for the username", () => {
    renderHook(() => useUserGET("john"));

    const fn = vi.mocked(useRemoteCall).mock.calls[0][0];
    fn();

    expect(getUser).toHaveBeenCalledWith("john");
});
