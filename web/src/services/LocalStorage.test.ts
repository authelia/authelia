// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { getLocalStorage, localStorageAvailable, setLocalStorage } from "@services/LocalStorage";

const mockLocalStorage = {
    getItem: vi.fn(),
    removeItem: vi.fn(),
    setItem: vi.fn(),
};

beforeEach(() => {
    mockLocalStorage.getItem.mockReset();
    mockLocalStorage.setItem.mockReset();
    mockLocalStorage.removeItem.mockReset();
    vi.stubGlobal("localStorage", mockLocalStorage);
});

afterEach(() => {
    vi.unstubAllGlobals();
});

it("reports localStorage as available", () => {
    expect(localStorageAvailable()).toBe(true);
});

it("gets value from localStorage", () => {
    mockLocalStorage.getItem.mockReturnValue("value");
    expect(getLocalStorage("key")).toBe("value");
});

it("returns null from localStorage when key is absent", () => {
    mockLocalStorage.getItem.mockReturnValue(null);
    expect(getLocalStorage("key")).toBeNull();
});

it("sets value in localStorage", () => {
    expect(setLocalStorage("key", "value")).toBe(true);
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith("key", "value");
});

describe("when localStorage is unusable", () => {
    beforeEach(() => {
        vi.resetModules();
    });

    it("reports it as unavailable when writes throw", async () => {
        vi.stubGlobal("localStorage", {
            getItem: vi.fn(),
            removeItem: vi.fn(),
            setItem: vi.fn(() => {
                throw new Error("quota exceeded");
            }),
        });

        const mod = await import("@services/LocalStorage");

        expect(mod.localStorageAvailable()).toBe(false);
        expect(mod.getLocalStorage("key")).toBeNull();
        expect(mod.setLocalStorage("key", "value")).toBe(false);
    });

    it("reports it as unavailable when the global is null", async () => {
        vi.stubGlobal("localStorage", null);

        const mod = await import("@services/LocalStorage");

        expect(mod.localStorageAvailable()).toBe(false);
    });

    it("caches the availability check", async () => {
        const setItem = vi.fn();
        vi.stubGlobal("localStorage", { getItem: vi.fn(), removeItem: vi.fn(), setItem });

        const mod = await import("@services/LocalStorage");

        expect(mod.localStorageAvailable()).toBe(true);
        expect(mod.localStorageAvailable()).toBe(true);
        expect(setItem).toHaveBeenCalledTimes(1);
    });
});
