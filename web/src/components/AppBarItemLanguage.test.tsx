import { fireEvent, render, screen } from "@testing-library/react";

import AppBarItemLanguage from "@components/AppBarItemLanguage";
import "@i18n/index";
import { Language } from "@models/LocaleInformation";

const mockOnChange = vi.fn();

const mockOf = vi.fn((locale: string) => {
    if (locale === "en") return "English";
    if (locale === "es") return "Spanish";
    if (locale === "fr") return "French";
    if (locale === "fr-CA") return "French (Canada)";
    if (locale === "fr-FR") return "French (France)";
    if (locale === "unknown") return "";
    return locale;
});

const OriginalDisplayNames = Intl.DisplayNames;

beforeAll(() => {
    (Intl as any).DisplayNames = class MockDisplayNames {
        of = mockOf;
        constructor() {}
    };
});

afterAll(() => {
    (Intl as any).DisplayNames = OriginalDisplayNames;
});

beforeEach(() => {
    vi.clearAllMocks();
    vi.spyOn(console, "error").mockImplementation(() => {});
});

const mockLanguages: Language[] = [
    { display: "English", fallbacks: [], locale: "en", namespaces: [] },
    { display: "Spanish", fallbacks: [], locale: "es", namespaces: [] },
    { display: "French", fallbacks: [], locale: "fr", namespaces: [] },
    { display: "French (Canada)", fallbacks: [], locale: "fr-CA", namespaces: [], parent: "fr" },
    { display: "French (France)", fallbacks: [], locale: "fr-FR", namespaces: [], parent: "fr" },
    { display: "Deutsch", fallbacks: [], locale: "de", namespaces: [] },
    { display: "Deutsch (Schweiz)", fallbacks: [], locale: "de-CH", namespaces: [], parent: "de" },
    { display: "", fallbacks: [], locale: "sc", namespaces: [] },
    { display: "", fallbacks: [], locale: "unknown", namespaces: [] },
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

it("resolves current locale from child locales", () => {
    render(<AppBarItemLanguage localeCurrent="fr-CA" localeList={mockLanguages} onChange={mockOnChange} />);
    expect(screen.getByText("French (Canada)")).toBeInTheDocument();
});

it("handles single child locale by promoting child locale", () => {
    render(<AppBarItemLanguage localeCurrent="de-CH" localeList={mockLanguages} onChange={mockOnChange} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(screen.getByText("Deutsch (de-CH)")).toBeInTheDocument();
});

it("returns null when current locale is not in list", () => {
    render(<AppBarItemLanguage localeCurrent="ja" localeList={mockLanguages} onChange={mockOnChange} />);
});

it("collapses and expands via icon click", () => {
    render(<AppBarItemLanguage localeCurrent="en" localeList={mockLanguages} onChange={mockOnChange} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    const frenchItem = screen.getByText("French (fr)");
    fireEvent.click(frenchItem);
    const expandLess = screen.getByTestId("ExpandLessIcon");
    fireEvent.click(expandLess);
    expect(screen.getByTestId("ExpandMoreIcon")).toBeInTheDocument();
    const expandMore = screen.getByTestId("ExpandMoreIcon");
    fireEvent.click(expandMore);
    expect(screen.getByTestId("ExpandLessIcon")).toBeInTheDocument();
});
