import { renderHook } from "@testing-library/react";

import { useAllGroupsGET } from "@hooks/GroupManagement";
import { useRemoteCall } from "@hooks/RemoteCall";
import { getAllGroups } from "@services/GroupManagement";

vi.mock("@hooks/RemoteCall", () => ({
    useRemoteCall: vi.fn(),
}));

it("calls useRemoteCall with getAllGroups", () => {
    vi.mocked(useRemoteCall).mockReturnValue("result" as any);

    const { result } = renderHook(() => useAllGroupsGET());

    expect(useRemoteCall).toHaveBeenCalledWith(getAllGroups);
    expect(result.current).toBe("result");
});
