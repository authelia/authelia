import { renderHook } from "@testing-library/react";

import { useRemoteCall } from "@hooks/RemoteCall";
import { useAutheliaState } from "@hooks/State";
import { getState } from "@services/State";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("calls useRemoteCall with getState", () => {
    (useRemoteCall as any).mockReturnValue("stateResult");
    const { result } = renderHook(() => useAutheliaState());
    expect(useRemoteCall).toHaveBeenCalledWith(getState);
    expect(result.current).toBe("stateResult");
});
