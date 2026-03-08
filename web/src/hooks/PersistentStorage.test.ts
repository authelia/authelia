import { act, renderHook } from "@testing-library/react";

import { usePersistentStorageValue } from "@hooks/PersistentStorage";

const mockLocalStorage = {
    getItem: vi.fn(),
    removeItem: vi.fn(),
    setItem: vi.fn(),
};

vi.stubGlobal("localStorage", mockLocalStorage);

beforeEach(() => {
    mockLocalStorage.getItem.mockReset();
    mockLocalStorage.setItem.mockReset();
    mockLocalStorage.removeItem.mockReset();
});

it("returns initial value when storage is empty", () => {
    mockLocalStorage.getItem.mockReturnValue(null);
    const { result } = renderHook(() => usePersistentStorageValue("key", "default"));
    expect(result.current[0]).toBe("default");
});

it("returns parsed value from storage", () => {
    mockLocalStorage.getItem.mockReturnValue('"stored"');
    const { result } = renderHook(() => usePersistentStorageValue("key", "default"));
    expect(result.current[0]).toBe("stored");
});

it("returns raw string when storage value is not valid JSON", () => {
    mockLocalStorage.getItem.mockReturnValue("plain string");
    const { result } = renderHook(() => usePersistentStorageValue("key", "default"));
    expect(result.current[0]).toBe("plain string");
});

it("returns null for stored null string", () => {
    mockLocalStorage.getItem.mockReturnValue("null");
    const { result } = renderHook(() => usePersistentStorageValue("key", "default"));
    expect(result.current[0]).toBe("default");
});

it("returns initial value for stored undefined string", () => {
    mockLocalStorage.getItem.mockReturnValue("undefined");
    const { result } = renderHook(() => usePersistentStorageValue("key", "default"));
    expect(result.current[0]).toBe("default");
});

it("merges object values from storage with initial value", () => {
    mockLocalStorage.getItem.mockReturnValue('{"b":2}');
    const { result } = renderHook(() => usePersistentStorageValue("key", { a: 1, b: 0 }));
    expect(result.current[0]).toEqual({ a: 1, b: 2 });
});

it("stores value to localStorage on update", () => {
    mockLocalStorage.getItem.mockReturnValue(null);
    const { result } = renderHook(() => usePersistentStorageValue("key", "default"));
    act(() => {
        result.current[1]("new value");
    });
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith("key", JSON.stringify("new value"));
});

it("removes item from localStorage when set to undefined", () => {
    mockLocalStorage.getItem.mockReturnValue('"existing"');
    const { result } = renderHook(() => usePersistentStorageValue<string | undefined>("key", "default"));
    act(() => {
        result.current[1](undefined);
    });
    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith("key");
});
