// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { renderHook } from "@testing-library/react";

import { useRemoteCall } from "@hooks/RemoteCall";
import { useUserSessionElevation } from "@hooks/UserSessionElevation";
import { getUserSessionElevation } from "@services/UserSessionElevation";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("calls useRemoteCall with getUserSessionElevation", () => {
    (useRemoteCall as any).mockReturnValue("elevationResult");
    const { result } = renderHook(() => useUserSessionElevation());
    expect(useRemoteCall).toHaveBeenCalledWith(getUserSessionElevation);
    expect(result.current).toBe("elevationResult");
});
