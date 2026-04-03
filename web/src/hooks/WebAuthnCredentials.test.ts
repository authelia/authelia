import { renderHook } from "@testing-library/react";

import { useRemoteCall } from "@hooks/RemoteCall";
import { useUserWebAuthnCredentials } from "@hooks/WebAuthnCredentials";
import { getUserWebAuthnCredentials } from "@services/UserWebAuthnCredentials";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("calls useRemoteCall with getUserWebAuthnCredentials", () => {
    (useRemoteCall as any).mockReturnValue("credentialsResult");
    const { result } = renderHook(() => useUserWebAuthnCredentials());
    expect(useRemoteCall).toHaveBeenCalledWith(getUserWebAuthnCredentials);
    expect(result.current).toBe("credentialsResult");
});
