import { act, fireEvent, render, screen } from "@testing-library/react";

import LocalStorageMethodContextProvider, { useLocalStorageMethodContext } from "@contexts/LocalStorageMethodContext";
import { SecondFactorMethod } from "@models/Methods";

vi.mock("@services/LocalStorage", () => ({
    localStorageAvailable: vi.fn(() => true),
}));

vi.mock("@constants/LocalStorage", () => ({
    LocalStorageSecondFactorMethod: "method",
}));

vi.mock("@services/UserInfo", () => ({
    isMethod2FA: vi.fn((value) => value === "totp" || value === "webauthn"),
    Method2FA: {},
    toMethod2FA: vi.fn((value) => value),
    toSecondFactorMethod: vi.fn((value) => value),
}));

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

const TestComponent = () => {
    const { localStorageMethod, localStorageMethodAvailable, setLocalStorageMethod } = useLocalStorageMethodContext();
    return (
        <div>
            <span>{localStorageMethod || "none"}</span>
            <span>{localStorageMethodAvailable ? "available" : "not available"}</span>
            <button onClick={() => setLocalStorageMethod("totp" as unknown as SecondFactorMethod)}>Set TOTP</button>
            <button onClick={() => setLocalStorageMethod(undefined)}>Clear</button>
        </div>
    );
};

it("renders without crashing", () => {
    render(
        <LocalStorageMethodContextProvider>
            <div>Test</div>
        </LocalStorageMethodContextProvider>,
    );
});

it("loads method from storage on mount", () => {
    mockLocalStorage.getItem.mockReturnValue("totp");
    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    expect(screen.getByText("totp")).toBeInTheDocument();
});

it("sets method in storage", () => {
    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    fireEvent.click(screen.getByText("Set TOTP"));
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith("method", "totp");
});

it("removes method from storage when set to undefined", async () => {
    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    await act(async () => {
        fireEvent.click(screen.getByText("Clear"));
    });
    expect(mockLocalStorage.removeItem).toHaveBeenCalledWith("method");
});

it("handles storage event for method change", async () => {
    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "method",
        newValue: "webauthn",
    });
    await act(async () => {
        window.dispatchEvent(event);
    });
    expect(await screen.findByText("webauthn")).toBeInTheDocument();
});

it("handles storage event with empty newValue", async () => {
    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "method",
        newValue: "",
    });
    await act(async () => {
        window.dispatchEvent(event);
    });
    expect(await screen.findByText("none")).toBeInTheDocument();
});

it("ignores storage event for different key", async () => {
    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "other",
        newValue: "totp",
    });
    await act(async () => {
        window.dispatchEvent(event);
    });
    expect(screen.getByText("none")).toBeInTheDocument();
});

it("handles localStorage being unavailable", async () => {
    const { localStorageAvailable } = await import("@services/LocalStorage");
    vi.mocked(localStorageAvailable).mockReturnValue(false);

    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    expect(screen.getByText("none")).toBeInTheDocument();
    expect(screen.getByText("not available")).toBeInTheDocument();

    vi.mocked(localStorageAvailable).mockReturnValue(true);
});

it("throws error if used outside provider", () => {
    expect(() => render(<TestComponent />)).toThrow(
        "useLocalStorageMethodContext must be used within a LocalStorageMethodContextProvider",
    );
});
