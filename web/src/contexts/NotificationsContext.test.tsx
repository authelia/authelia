import { ReactNode } from "react";

import { act, renderHook } from "@testing-library/react";

import NotificationsContextProvider, { useNotifications } from "@contexts/NotificationsContext";

const mockSetNotification = vi.fn();

beforeEach(() => {
    mockSetNotification.mockReset();
});

const wrapper = ({ children }: { children: ReactNode }) => (
    <NotificationsContextProvider>{children}</NotificationsContextProvider>
);

it("creates info notification with default timeout", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createInfoNotification("info message");
    });
    expect(mockSetNotification).toHaveBeenCalledWith({
        level: "info",
        message: "info message",
        timeout: 5,
    });
});

it("creates success notification", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createSuccessNotification("success message");
    });
    expect(mockSetNotification).toHaveBeenCalledWith({
        level: "success",
        message: "success message",
        timeout: 5,
    });
});

it("creates warning notification", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createWarnNotification("warn message");
    });
    expect(mockSetNotification).toHaveBeenCalledWith({
        level: "warning",
        message: "warn message",
        timeout: 5,
    });
});

it("creates error notification", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createErrorNotification("error message");
    });
    expect(mockSetNotification).toHaveBeenCalledWith({
        level: "error",
        message: "error message",
        timeout: 5,
    });
});

it("creates notification with custom timeout", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createInfoNotification("message", 10);
    });
    expect(mockSetNotification).toHaveBeenCalledWith({
        level: "info",
        message: "message",
        timeout: 10,
    });
});

it("resets notification", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.resetNotification();
    });
    expect(mockSetNotification).toHaveBeenCalledWith(null);
});

it("reports inactive when notification is null", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    expect(result.current.isActive).toBe(false);
});

/*
it("reports active when notification exists", () => {
    const activeWrapper = ({ children }: { children: ReactNode }) => (
        <NotificationsContext.Provider
            value={{
                notification: { level: "info", message: "test", timeout: 5 },
                setNotification: mockSetNotification,
            }}
        >
            {children}
        </NotificationsContext.Provider>
    );
    const { result } = renderHook(() => useNotifications(), { wrapper: activeWrapper });
    expect(result.current.isActive).toBe(true);
});
 */
