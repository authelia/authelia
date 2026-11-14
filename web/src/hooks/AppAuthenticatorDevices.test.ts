// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { renderHook } from "@testing-library/react";

import { useAppAuthenticatorDevices } from "@hooks/AppAuthenticatorDevices";
import { useRemoteCall } from "@hooks/RemoteCall";
import { getAppAuthenticatorDevices } from "@services/AppAuthenticator";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("calls useRemoteCall with getAppAuthenticatorDevices", () => {
    (useRemoteCall as any).mockReturnValue("devicesResult");
    const { result } = renderHook(() => useAppAuthenticatorDevices());
    expect(useRemoteCall).toHaveBeenCalledWith(getAppAuthenticatorDevices);
    expect(result.current).toBe("devicesResult");
});
