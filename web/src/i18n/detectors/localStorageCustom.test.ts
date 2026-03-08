import LocalStorageCustomDetector from "@i18n/detectors/localStorageCustom";
import { getLocalStorage, localStorageAvailable } from "@services/LocalStorage";

vi.mock("@constants/LocalStorage", () => ({
    LocalStorageLanguagePreference: "language",
}));

vi.mock("@services/LocalStorage", () => ({
    getLocalStorage: vi.fn(),
    localStorageAvailable: vi.fn(() => true),
}));

beforeEach(() => {
    vi.mocked(getLocalStorage).mockReset();
    vi.mocked(localStorageAvailable).mockReset().mockReturnValue(true);
});

it("returns language from localStorage when available", () => {
    vi.mocked(getLocalStorage).mockReturnValue("fr");
    const result = LocalStorageCustomDetector.lookup({ lookupLocalStorage: "language" } as any);
    expect(result).toBe("fr");
    expect(getLocalStorage).toHaveBeenCalledWith("language");
});

it("returns undefined when lookupLocalStorage is not set", () => {
    const result = LocalStorageCustomDetector.lookup({} as any);
    expect(result).toBeUndefined();
});

it("returns undefined when localStorage is not available", () => {
    vi.mocked(localStorageAvailable).mockReturnValue(false);
    const result = LocalStorageCustomDetector.lookup({ lookupLocalStorage: "language" } as any);
    expect(result).toBeUndefined();
});

it("returns undefined when localStorage value is empty", () => {
    vi.mocked(getLocalStorage).mockReturnValue("");
    const result = LocalStorageCustomDetector.lookup({ lookupLocalStorage: "language" } as any);
    expect(result).toBeUndefined();
});

it("returns undefined when localStorage value is null", () => {
    vi.mocked(getLocalStorage).mockReturnValue(null);
    const result = LocalStorageCustomDetector.lookup({ lookupLocalStorage: "language" } as any);
    expect(result).toBeUndefined();
});

it("has the correct detector name", () => {
    expect(LocalStorageCustomDetector.name).toBe("localStorageCustom");
});
