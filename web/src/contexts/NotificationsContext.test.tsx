import { ReactNode } from "react";

import { act, renderHook } from "@testing-library/react";

import NotificationsContextProvider, { useNotifications } from "@contexts/NotificationsContext";

const wrapper = ({ children }: { children: ReactNode }) => (
    <NotificationsContextProvider>{children}</NotificationsContextProvider>
);

it("creates info notification with default timeout", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createInfoNotification("info message");
    });
    expect(result.current.notification).toEqual({
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
    expect(result.current.notification).toEqual({
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
    expect(result.current.notification).toEqual({
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
    expect(result.current.notification).toEqual({
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
    expect(result.current.notification).toEqual({
        level: "info",
        message: "message",
        timeout: 10,
    });
});

it("shows notification with showNotification", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.showNotification("success", "direct message", 7);
    });
    expect(result.current.notification).toEqual({
        level: "success",
        message: "direct message",
        timeout: 7,
    });
});

it("resets notification", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createInfoNotification("message");
    });
    expect(result.current.notification).not.toBeNull();
    act(() => {
        result.current.resetNotification();
    });
    expect(result.current.notification).toBeNull();
});

it("reports inactive when notification is null", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    expect(result.current.isActive).toBe(false);
});

it("reports active when notification exists", () => {
    const { result } = renderHook(() => useNotifications(), { wrapper });
    act(() => {
        result.current.createInfoNotification("test");
    });
    expect(result.current.isActive).toBe(true);
});

it("throws when used outside provider", () => {
    expect(() => {
        renderHook(() => useNotifications());
    }).toThrow("useNotifications must be used within a NotificationsProvider");
});
