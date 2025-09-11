import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import LanguageContextProvider, { useLanguageContext } from "@contexts/LanguageContext";

const mockI18n: any = {
    resolvedLanguage: "en",
    language: "en",
    changeLanguage: vi.fn(() => Promise.resolve()),
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

it("updates locale", () => {
    render(
        <LanguageContextProvider i18n={mockI18n}>
            <TestComponent />
        </LanguageContextProvider>,
    );
    const button = screen.getByRole("button");
    button.click();
    expect(mockI18n.changeLanguage).toHaveBeenCalledWith("fr");
});

it("handles storage event for language change", () => {
    render(
        <LanguageContextProvider i18n={mockI18n}>
            <TestComponent />
        </LanguageContextProvider>,
    );
    const event = new StorageEvent("storage", {
        key: "language",
        newValue: "fr",
    });
    window.dispatchEvent(event);
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
