import React from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import AppBarItemLanguage from "@components/AppBarItemLanguage";
import { Language } from "@models/LocaleInformation";

const mockOnChange = vi.fn();

const mockOf = vi.fn((locale: string) => {
    if (locale === "en") return "English";
    if (locale === "es") return "Spanish";
    if (locale === "fr") return "French";
    if (locale === "fr-CA") return "French (Canada)";
    if (locale === "fr-FR") return "French (France)";
    return locale;
});
const originalIntl = global.Intl;

beforeAll(() => {
    vi.stubGlobal("Intl", {
        DisplayNames: vi.fn(() => ({ of: mockOf })),
    });
});

afterAll(() => {
    vi.unstubAllGlobals();
    global.Intl = originalIntl;
});

beforeEach(() => {
    vi.clearAllMocks();
});

const mockLanguages: Language[] = [
    { display: "English", locale: "en", fallbacks: [], namespaces: [] },
    { display: "Spanish", locale: "es", fallbacks: [], namespaces: [] },
    { display: "French", locale: "fr", fallbacks: [], namespaces: [] },
    { display: "French (Canada)", locale: "fr-CA", parent: "fr", fallbacks: [], namespaces: [] },
    { display: "French (France)", locale: "fr-FR", parent: "fr", fallbacks: [], namespaces: [] },
    { display: "", locale: "sc", fallbacks: [], namespaces: [] },
    { display: "", locale: "unknown", fallbacks: [], namespaces: [] },
];

it("renders without crashing", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
});

it("does not render when props missing", () => {
    const { container } = render(<AppBarItemLanguage />);
    expect(container).toBeEmptyDOMElement();
});

it("renders button with current language", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
    expect(screen.getByText("English")).toBeInTheDocument();
});

it("opens menu on button click", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(screen.getByText("English (en)")).toBeInTheDocument();
    expect(screen.getByText("Spanish (es)")).toBeInTheDocument();
});

it("changes language on selection", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    const spanishItem = screen.getByText("Spanish (es)");
    fireEvent.click(spanishItem);
    expect(mockOnChange).toHaveBeenCalledWith("es");
});

it("expands and collapses parent with children on click", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    const frenchItem = screen.getByText("French (fr)");
    fireEvent.click(frenchItem);
    expect(screen.getByTestId("ExpandLessIcon")).toBeInTheDocument();
    fireEvent.click(frenchItem);
    expect(screen.getByTestId("ExpandMoreIcon")).toBeInTheDocument();
});

it("changes language on child selection", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    const frenchItem = screen.getByText("French (fr)");
    fireEvent.click(frenchItem);
    const childItem = screen.getByText("French (Canada) (fr-CA)");
    fireEvent.click(childItem);
    expect(mockOnChange).toHaveBeenCalledWith("fr-CA");
});

it("uses fallback for unknown locale", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(screen.getByText("Basa Sunda (sc)")).toBeInTheDocument();
});
