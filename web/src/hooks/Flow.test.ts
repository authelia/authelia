import { renderHook } from "@testing-library/react";
import { useSearchParams } from "react-router-dom";

import { useFlow, useFlowPresent } from "@hooks/Flow";

vi.mock("react-router-dom", () => ({
    useSearchParams: vi.fn(),
}));

vi.mock("@constants/SearchParams", () => ({
    Flow: "flow",
    FlowID: "flow_id",
    RedirectionURL: "rd",
    SubFlow: "subflow",
}));

it("returns flow parameters when present", () => {
    const mockSearchParams = new URLSearchParams("flow=login&flow_id=123&subflow=auth");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useFlow());
    expect(result.current).toEqual({ flow: "login", id: "123", subflow: "auth" });
});

it("returns undefined values when parameters are absent", () => {
    const mockSearchParams = new URLSearchParams("");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useFlow());
    expect(result.current).toEqual({ flow: undefined, id: undefined, subflow: undefined });
});

it("returns true when flow is present", () => {
    const mockSearchParams = new URLSearchParams("flow=login");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useFlowPresent());
    expect(result.current).toBe(true);
});

it("returns true when redirection is present", () => {
    const mockSearchParams = new URLSearchParams("rd=https://example.com");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useFlowPresent());
    expect(result.current).toBe(true);
});

it("returns false when neither flow nor redirection is present", () => {
    const mockSearchParams = new URLSearchParams("");
    (useSearchParams as any).mockReturnValue([mockSearchParams]);
    const { result } = renderHook(() => useFlowPresent());
    expect(result.current).toBe(false);
});
