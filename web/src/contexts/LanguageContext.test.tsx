import { act, fireEvent, render, screen } from "@testing-library/react";

import LanguageContextProvider, { useLanguageContext } from "@contexts/LanguageContext";

const mockI18n: any = {
    changeLanguage: vi.fn(() => Promise.resolve()),
    language: "en",
    resolvedLanguage: "en",
};

vi.mock("@services/LocalStorage", () => ({
    setLocalStorage: vi.fn(),
}));

vi.mock("@constants/LocalStorage", () => ({
    LocalStorageLanguagePreference: "language",
}));

beforeEach(() => {
    mockI18n.changeLanguage.mockClear();
});

const TestComponent = () => {
    const { locale, setLocale } = useLanguageContext();
    return (
        <div>
            <span>{locale}</span>
            <button onClick={() => setLocale("fr")}>Change</button>
        </div>
    );
};

it("renders without crashing", () => {
    render(
        <LanguageContextProvider i18n={mockI18n}>
            <div>Test</div>
        </LanguageContextProvider>,
    );
});

it("updates locale", async () => {
    render(
        <LanguageContextProvider i18n={mockI18n}>
            <TestComponent />
        </LanguageContextProvider>,
    );
    await act(async () => {
        fireEvent.click(screen.getByRole("button"));
    });
    expect(mockI18n.changeLanguage).toHaveBeenCalledWith("fr");
});

it("handles storage event for language change", async () => {
    render(
        <LanguageContextProvider i18n={mockI18n}>
            <TestComponent />
        </LanguageContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "language",
        newValue: "fr",
    });
    await act(async () => {
        window.dispatchEvent(event);
    });
    expect(mockI18n.changeLanguage).toHaveBeenCalledWith("fr");
});

it("ignores storage event for different key", () => {
    render(
        <LanguageContextProvider i18n={mockI18n}>
            <TestComponent />
        </LanguageContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "other",
        newValue: "fr",
    });
    window.dispatchEvent(event);
    expect(mockI18n.changeLanguage).not.toHaveBeenCalledWith("fr");
});

it("ignores storage event with empty value", () => {
    render(
        <LanguageContextProvider i18n={mockI18n}>
            <TestComponent />
        </LanguageContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "language",
        newValue: "",
    });
    window.dispatchEvent(event);
    expect(mockI18n.changeLanguage).not.toHaveBeenCalledWith("");
});

it("throws error if used outside provider", () => {
    expect(() => render(<TestComponent />)).toThrow("useLanguageContext must be used within a LanguageContextProvider");
});
