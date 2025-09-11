import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

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
    toMethod2FA: vi.fn((value) => value),
    toSecondFactorMethod: vi.fn((value) => value),
    Method2FA: {},
}));

const mockLocalStorage = {
    getItem: vi.fn(),
    setItem: vi.fn(),
    removeItem: vi.fn(),
};

vi.stubGlobal("localStorage", mockLocalStorage);

const TestComponent = () => {
    const { localStorageMethod, setLocalStorageMethod, localStorageMethodAvailable } = useLocalStorageMethodContext();
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
    const button = screen.getByText("Set TOTP");
    button.click();
    expect(mockLocalStorage.setItem).toHaveBeenCalledWith("method", "totp");
});

it("removes method from storage when set to undefined", () => {
    render(
        <LocalStorageMethodContextProvider>
            <TestComponent />
        </LocalStorageMethodContextProvider>,
    );
    const button = screen.getByText("Clear");
    button.click();
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
    window.dispatchEvent(event);
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
    window.dispatchEvent(event);
    expect(await screen.findByText("none")).toBeInTheDocument();
});

it("throws error if used outside provider", () => {
    expect(() => render(<TestComponent />)).toThrow(
        "useLocalStorageMethodContext must be used within a LocalStorageMethodContextProvider",
    );
});
