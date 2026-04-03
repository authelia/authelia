import { renderHook } from "@testing-library/react";

import { useConfiguration } from "@hooks/Configuration";
import { useRemoteCall } from "@hooks/RemoteCall";
import { getConfiguration } from "@services/Configuration";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("calls useRemoteCall with getConfiguration", () => {
    (useRemoteCall as any).mockReturnValue("configResult");
    const { result } = renderHook(() => useConfiguration());
    expect(useRemoteCall).toHaveBeenCalledWith(getConfiguration);
    expect(result.current).toBe("configResult");
});
