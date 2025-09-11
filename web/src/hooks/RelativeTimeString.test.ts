import { renderHook, waitFor } from "@testing-library/react";
import { vi } from "vitest";

import { getRelativeTimeString, useRelativeTime } from "@hooks/RelativeTimeString";

vi.mock("i18next", () => ({
    default: {
        languages: ["en"],
    },
}));

const mockRelativeTimeFormat = vi.fn((value, unit) => `${Math.abs(value)} ${unit} ago`);
vi.spyOn(Intl, "RelativeTimeFormat").mockImplementation(
    () =>
        ({
            format: mockRelativeTimeFormat,
        }) as any,
);

it("returns seconds ago for less than a minute", () => {
    const date = new Date(Date.now() - 30 * 1000); // 30 seconds ago
    const result = getRelativeTimeString(date);
    expect(mockRelativeTimeFormat).toHaveBeenCalledWith(0, "seconds");
});

it("returns minutes ago for less than an hour", () => {
    const date = new Date(Date.now() - 5 * 60 * 1000); // 5 minutes ago
    const result = getRelativeTimeString(date);
    expect(mockRelativeTimeFormat).toHaveBeenCalledWith(-5, "minutes");
});

it("returns hours ago for less than a day", () => {
    const date = new Date(Date.now() - 2 * 60 * 60 * 1000); // 2 hours ago
    const result = getRelativeTimeString(date);
    expect(mockRelativeTimeFormat).toHaveBeenCalledWith(-2, "hours");
});

it("returns days ago for less than a month", () => {
    const date = new Date(Date.now() - 3 * 24 * 60 * 60 * 1000); // 3 days ago
    const result = getRelativeTimeString(date);
    expect(mockRelativeTimeFormat).toHaveBeenCalledWith(-3, "days");
});

it("returns months ago for less than a year", () => {
    const date = new Date(Date.now() - 2 * 30 * 24 * 60 * 60 * 1000); // 2 months ago
    const result = getRelativeTimeString(date);
    expect(mockRelativeTimeFormat).toHaveBeenCalledWith(-2, "months");
});

it("returns years ago for more than a year", () => {
    const date = new Date(Date.now() - 2 * 365 * 24 * 60 * 60 * 1000); // 2 years ago
    const result = getRelativeTimeString(date);
    expect(mockRelativeTimeFormat).toHaveBeenCalledWith(-2, "years");
});

it("returns 0 seconds ago for current date", () => {
    const date = new Date(Date.now()); // now
    const result = getRelativeTimeString(date);
    expect(result).toBe("0 seconds ago");
});
