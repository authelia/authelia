import { act, renderHook } from "@testing-library/react";

import { getRelativeTimeString, useRelativeTime } from "@hooks/RelativeTimeString";

vi.mock("i18next", () => ({
    default: {
        languages: ["en"],
    },
}));

const mockFormat = vi.fn((value: number, unit: Intl.RelativeTimeFormatUnit) => {
    return `${Math.abs(value)} ${unit} ago`;
});

const OriginalRelativeTimeFormat = Intl.RelativeTimeFormat;

beforeAll(() => {
    (Intl as any).RelativeTimeFormat = class MockRelativeTimeFormat {
        format = mockFormat;
        constructor() {}
    };
});

afterAll(() => {
    (Intl as any).RelativeTimeFormat = OriginalRelativeTimeFormat;
});

beforeEach(() => {
    mockFormat.mockClear();
});

it("returns seconds ago for less than a minute", () => {
    const date = new Date(Date.now() - 30 * 1000);
    getRelativeTimeString(date);
    expect(mockFormat).toHaveBeenCalledWith(0, "seconds");
});

it("returns minutes ago for less than an hour", () => {
    const date = new Date(Date.now() - 5 * 60 * 1000);
    getRelativeTimeString(date);
    expect(mockFormat).toHaveBeenCalledWith(-5, "minutes");
});

it("returns hours ago for less than a day", () => {
    const date = new Date(Date.now() - 2 * 60 * 60 * 1000);
    getRelativeTimeString(date);
    expect(mockFormat).toHaveBeenCalledWith(-2, "hours");
});

it("returns days ago for less than a month", () => {
    const date = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000);
    getRelativeTimeString(date);
    expect(mockFormat).toHaveBeenCalledWith(-3, "days");
});

it("returns months ago for less than a year", () => {
    const date = new Date(Date.now() - 2 * 30 * 24 * 60 * 60 * 1000);
    getRelativeTimeString(date);
    expect(mockFormat).toHaveBeenCalledWith(-2, "months");
});

it("returns years ago for more than a year", () => {
    const date = new Date(Date.now() - 2 * 365 * 24 * 60 * 60 * 1000);
    getRelativeTimeString(date);
    expect(mockFormat).toHaveBeenCalledWith(-2, "years");
});

it("returns 0 seconds ago for current date", () => {
    const date = new Date(Date.now());
    const result = getRelativeTimeString(date);
    expect(result).toBe("0 seconds ago");
});

it("returns never for exactly one year", () => {
    const date = new Date(Date.now() - 31536000 * 1000);
    const result = getRelativeTimeString(date);
    expect(result).toBe("never");
});

it("updates relative time on interval", () => {
    vi.useFakeTimers();
    const date = new Date(Date.now() - 30 * 1000);
    const { result } = renderHook(() => useRelativeTime(date));
    expect(result.current).toBe("0 seconds ago");

    mockFormat.mockClear();

    act(() => {
        vi.advanceTimersByTime(60000);
    });

    expect(mockFormat).toHaveBeenCalled();
    vi.useRealTimers();
});

it("cleans up interval on unmount", () => {
    vi.useFakeTimers();
    const date = new Date(Date.now() - 30 * 1000);
    const { unmount } = renderHook(() => useRelativeTime(date));

    unmount();
    mockFormat.mockClear();
    vi.advanceTimersByTime(120000);
    expect(mockFormat).not.toHaveBeenCalled();
    vi.useRealTimers();
});
