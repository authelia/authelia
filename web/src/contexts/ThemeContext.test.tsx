import { act, fireEvent, render, screen } from "@testing-library/react";

import ThemeContextProvider, { useThemeContext } from "@contexts/ThemeContext";

vi.mock("@constants/LocalStorage", () => ({
    LocalStorageThemeName: "theme",
}));

vi.mock("@services/LocalStorage", () => ({
    localStorageAvailable: vi.fn(() => true),
    setLocalStorage: vi.fn(),
}));

const mockLocalStorage = {
    getItem: vi.fn(),
    removeItem: vi.fn(),
    setItem: vi.fn(),
};

const mockMatchMedia = {
    addEventListener: vi.fn(),
    matches: false,
    removeEventListener: vi.fn(),
};

vi.stubGlobal("localStorage", mockLocalStorage);

beforeEach(() => {
    mockLocalStorage.getItem.mockReset();
    mockLocalStorage.setItem.mockReset();
    mockLocalStorage.removeItem.mockReset();
    mockMatchMedia.addEventListener.mockReset();
    mockMatchMedia.removeEventListener.mockReset();
    mockMatchMedia.matches = false;
});

vi.stubGlobal(
    "matchMedia",
    vi.fn(() => mockMatchMedia),
);

const TestComponent = () => {
    const { setThemeName, themeName } = useThemeContext();
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
    await act(async () => {
        fireEvent.click(screen.getByText("Set Dark"));
    });
    expect(await screen.findByText("dark")).toBeInTheDocument();
});

it("sets theme name to auto", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    await act(async () => {
        fireEvent.click(screen.getByText("Set Auto"));
    });
    expect(await screen.findByText("auto")).toBeInTheDocument();
});

it("sets theme name to oled", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    await act(async () => {
        fireEvent.click(screen.getByText("Set Oled"));
    });
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
    await act(async () => {
        globalThis.dispatchEvent(event);
    });
    expect(await screen.findByText("grey")).toBeInTheDocument();
});

it("handles storage event with empty newValue", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "theme",
        newValue: "",
    });
    await act(async () => {
        globalThis.dispatchEvent(event);
    });
    expect(screen.getByText("light")).toBeInTheDocument();
});

it("handles storage event with empty newValue falling back to stored theme", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    mockLocalStorage.getItem.mockReturnValue("dark");
    const event = new StorageEvent("storage", {
        key: "theme",
        newValue: "",
    });
    await act(async () => {
        globalThis.dispatchEvent(event);
    });
    expect(await screen.findByText("dark")).toBeInTheDocument();
});

it("ignores storage event for different key", async () => {
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    await act(async () => {
        fireEvent.click(screen.getByText("Set Dark"));
    });
    expect(await screen.findByText("dark")).toBeInTheDocument();
    const event = new StorageEvent("storage", {
        key: "other",
        newValue: "grey",
    });
    await act(async () => {
        globalThis.dispatchEvent(event);
    });
    expect(screen.getByText("dark")).toBeInTheDocument();
});

it("initializes with auto theme from storage", () => {
    mockLocalStorage.getItem.mockReturnValue("auto");
    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    expect(screen.getByText("auto")).toBeInTheDocument();
});

it("responds to media query change when auto theme is set", async () => {
    let mediaQueryCallback: ((_ev: MediaQueryListEvent) => void) | undefined;
    mockMatchMedia.addEventListener.mockImplementation((_event: string, cb: (_ev: MediaQueryListEvent) => void) => {
        mediaQueryCallback = cb;
    });

    render(
        <ThemeContextProvider>
            <TestComponent />
        </ThemeContextProvider>,
    );
    await act(async () => {
        fireEvent.click(screen.getByText("Set Auto"));
    });
    expect(await screen.findByText("auto")).toBeInTheDocument();
    expect(mockMatchMedia.addEventListener).toHaveBeenCalledWith("change", expect.any(Function));

    await act(async () => {
        mediaQueryCallback!({ matches: true } as MediaQueryListEvent);
    });
});

it("throws error if used outside provider", () => {
    expect(() => render(<TestComponent />)).toThrow("useThemeContext must be used within a ThemeContextProvider");
});
