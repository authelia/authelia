import { renderHook } from "@testing-library/react";

import { useRemoteCall } from "@hooks/RemoteCall";
import { useUserInfoTOTPConfiguration, useUserInfoTOTPConfigurationOptional } from "@hooks/UserInfoTOTPConfiguration";
import {
    getUserInfoTOTPConfiguration,
    getUserInfoTOTPConfigurationOptional,
} from "@services/UserInfoTOTPConfiguration";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("calls useRemoteCall with getUserInfoTOTPConfiguration", () => {
    (useRemoteCall as any).mockReturnValue("totpResult");
    const { result } = renderHook(() => useUserInfoTOTPConfiguration());
    expect(useRemoteCall).toHaveBeenCalledWith(getUserInfoTOTPConfiguration);
    expect(result.current).toBe("totpResult");
});

it("calls useRemoteCall with getUserInfoTOTPConfigurationOptional", () => {
    (useRemoteCall as any).mockReturnValue("totpOptionalResult");
    const { result } = renderHook(() => useUserInfoTOTPConfigurationOptional());
    expect(useRemoteCall).toHaveBeenCalledWith(getUserInfoTOTPConfigurationOptional);
    expect(result.current).toBe("totpOptionalResult");
});
