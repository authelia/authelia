// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { renderHook } from "@testing-library/react";

import { useRemoteCall } from "@hooks/RemoteCall";
import { useGenerateRecoveryCodes, useRecoveryCodesStatus } from "@hooks/UserRecoveryCodes";
import { generateRecoveryCodes, getRecoveryCodesStatus } from "@services/UserRecoveryCodes";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

vi.mock("@services/UserRecoveryCodes", () => ({
    generateRecoveryCodes: vi.fn(),
    getRecoveryCodesStatus: vi.fn(),
}));

beforeEach(() => {
    vi.clearAllMocks();
});

it("calls useRemoteCall with getRecoveryCodesStatus", () => {
    vi.mocked(useRemoteCall).mockReturnValue("statusResult" as any);
    const { result } = renderHook(() => useRecoveryCodesStatus());
    expect(useRemoteCall).toHaveBeenCalledWith(getRecoveryCodesStatus);
    expect(result.current).toBe("statusResult");
});

it("calls useRemoteCall with generateRecoveryCodes and resets data on each call", () => {
    vi.mocked(useRemoteCall).mockReturnValue("generateResult" as any);
    const { result } = renderHook(() => useGenerateRecoveryCodes());
    expect(useRemoteCall).toHaveBeenCalledWith(generateRecoveryCodes, { resetDataOnCall: true });
    expect(result.current).toBe("generateResult");
});
