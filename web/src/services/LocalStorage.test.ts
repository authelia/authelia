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
