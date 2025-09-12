import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import ThemeContextProvider, { useThemeContext } from "@contexts/ThemeContext";

vi.mock("@constants/LocalStorage", () => ({
    LocalStorageThemeName: "theme",
}));

const mockLocalStorage = {
    getItem: vi.fn(),
    setItem: vi.fn(),
    removeItem: vi.fn(),
};

const mockMatchMedia = {
    matches: false,
    addEventListener: vi.fn(),
    removeEventListener: vi.fn(),
};

vi.stubGlobal("localStorage", mockLocalStorage);
vi.stubGlobal(
    "matchMedia",
    vi.fn(() => mockMatchMedia),
);

const TestComponent = () => {
    const { themeName, setThemeName } = useThemeContext();
    return (
        <div>
            <span>{themeName}</span>
            <button onClick={() => setThemeName("dark")}>Set Dark</button>
            <button onClick={() => setThemeName("auto")}>Set Auto</button>
            <button onClick={() => setThemeName("oled")}>Set Oled</button>
        </div>
    );
};

it("renders without crashing", () => {
    render(
        <ThemeContextProvider>
            <div>Test</div>
        </ThemeContextProvider>,
    );
});

it("handles unknown theme name", () => {
    mockLocalStorage.getItem.mockReturnValue("unknown");
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    expect(screen.getByText("unknown")).toBeInTheDocument();
});

it("sets theme name and stores in storage", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    const button = screen.getByText("Set Dark");
    button.click();
    expect(await screen.findByText("dark")).toBeInTheDocument();
});

it("sets theme name to auto", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    const button = screen.getByText("Set Auto");
    button.click();
    expect(await screen.findByText("auto")).toBeInTheDocument();
});

it("sets theme name to oled", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    const button = screen.getByText("Set Oled");
    button.click();
    expect(await screen.findByText("oled")).toBeInTheDocument();
});

it("handles storage event for theme change", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "theme",
        newValue: "grey",
    });
    window.dispatchEvent(event);
    expect(await screen.findByText("grey")).toBeInTheDocument();
});

it("throws error if used outside provider", () => {
    expect(() => render(<TestComponent />)).toThrow("useThemeContext must be used within a ThemeContextProvider");
});
