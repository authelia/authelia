import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";

import LoginLayout from "@layouts/LoginLayout";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { body: "body", icon: "icon", root: "root", rootContainer: "rootContainer" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@assets/images/user.svg?react", () => ({
    default: (props: any) => <svg data-testid="user-svg" {...props} />,
}));

vi.mock("@components/AppBarLoginPortal", () => ({
    default: (props: any) => (
        <div data-testid="app-bar" data-locale={props.localeCurrent}>
            <button data-testid="change-locale" onClick={() => props.onLocaleChange?.("fr")} />
        </div>
    ),
}));

vi.mock("@components/Brand", () => ({
    default: () => <div data-testid="brand" />,
}));

vi.mock("@components/PrivacyPolicyDrawer", () => ({
    default: () => <div data-testid="privacy-policy" />,
}));

vi.mock("@components/TypographyWithTooltip", () => ({
    default: (props: any) => <span data-testid={`typography-${props.variant}`}>{props.value}</span>,
}));

vi.mock("@constants/constants", () => ({
    EncodedName: [81, 88, 86, 48, 97, 71, 86, 115, 97, 87, 69, 61],
}));

const mockSetLocale = vi.fn();
vi.mock("@contexts/LanguageContext", () => ({
    useLanguageContext: () => ({ locale: "en", setLocale: mockSetLocale }),
}));

vi.mock("@services/LocaleInformation", () => ({
    getLocaleInformation: vi.fn().mockResolvedValue({ languages: [{ display: "English", locale: "en" }] }),
}));

vi.mock("@utils/Configuration", () => ({
    getLogoOverride: vi.fn(() => false),
}));

beforeEach(() => {
    mockSetLocale.mockReset();
});

afterEach(() => {
    vi.restoreAllMocks();
});

it("renders with default SVG logo", async () => {
    await act(async () => {
        render(<LoginLayout />);
    });

    expect(screen.getByTestId("user-svg")).toBeInTheDocument();
    expect(screen.getByTestId("app-bar")).toBeInTheDocument();
    expect(screen.getByTestId("brand")).toBeInTheDocument();
    expect(screen.getByTestId("privacy-policy")).toBeInTheDocument();
});

it("renders with image logo when override is enabled", async () => {
    const { getLogoOverride } = await import("@utils/Configuration");
    vi.mocked(getLogoOverride).mockReturnValue(true);

    await act(async () => {
        render(<LoginLayout />);
    });

    expect(screen.getByAltText("Logo")).toBeInTheDocument();
    expect(screen.queryByTestId("user-svg")).not.toBeInTheDocument();

    vi.mocked(getLogoOverride).mockReturnValue(false);
});

it("renders title and subtitle when provided", async () => {
    await act(async () => {
        render(<LoginLayout title="Test Title" subtitle="Test Subtitle" />);
    });

    expect(screen.getByTestId("typography-h5")).toHaveTextContent("Test Title");
    expect(screen.getByTestId("typography-h6")).toHaveTextContent("Test Subtitle");
});

it("does not render title or subtitle when not provided", async () => {
    await act(async () => {
        render(<LoginLayout />);
    });

    expect(screen.queryByTestId("typography-h5")).not.toBeInTheDocument();
    expect(screen.queryByTestId("typography-h6")).not.toBeInTheDocument();
});

it("renders children", async () => {
    await act(async () => {
        render(
            <LoginLayout>
                <div data-testid="child">Content</div>
            </LoginLayout>,
        );
    });

    expect(screen.getByTestId("child")).toBeInTheDocument();
});

it("sets the document title", async () => {
    document.title = "Sentinel Title";

    await act(async () => {
        render(<LoginLayout />);
    });

    expect(document.title).not.toBe("Sentinel Title");
    expect(document.title).toContain("Login");
});

it("calls setLocale when language is changed", async () => {
    await act(async () => {
        render(<LoginLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByTestId("change-locale"));
    });

    expect(mockSetLocale).toHaveBeenCalledWith("fr");
});

it("logs error when locale fetch fails", async () => {
    vi.spyOn(console, "error").mockImplementation(() => {});

    const { getLocaleInformation } = await import("@services/LocaleInformation");
    vi.mocked(getLocaleInformation).mockRejectedValueOnce(new Error("fetch failed"));

    render(<LoginLayout />);

    await waitFor(() => {
        expect(console.error).toHaveBeenCalledWith("could not get locale list:", expect.any(Error));
    });
});
