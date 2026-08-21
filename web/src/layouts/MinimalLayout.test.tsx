import { act, render, screen } from "@testing-library/react";

import MinimalLayout from "@layouts/MinimalLayout";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@assets/images/user.svg?react", () => ({
    default: (props: any) => <svg data-testid="user-svg" {...props} />,
}));

vi.mock("@components/AppBarLoginPortal", () => ({
    default: () => <div data-testid="app-bar" />,
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

vi.mock("@utils/Configuration", () => ({
    getLogoOverride: vi.fn(() => false),
}));

it("renders with default SVG logo", async () => {
    await act(async () => {
        render(<MinimalLayout />);
    });

    expect(screen.getByTestId("user-svg")).toBeInTheDocument();
    expect(screen.getByTestId("app-bar")).toBeInTheDocument();
    expect(screen.getByTestId("privacy-policy")).toBeInTheDocument();
});

it("renders with image logo when override is enabled", async () => {
    const { getLogoOverride } = await import("@utils/Configuration");
    vi.mocked(getLogoOverride).mockReturnValue(true);

    await act(async () => {
        render(<MinimalLayout />);
    });

    expect(screen.getByAltText("Logo")).toBeInTheDocument();
    expect(screen.queryByTestId("user-svg")).not.toBeInTheDocument();

    vi.mocked(getLogoOverride).mockReturnValue(false);
});

it("renders title when provided", async () => {
    await act(async () => {
        render(<MinimalLayout title="Test Title" />);
    });

    expect(screen.getByTestId("typography-h5")).toHaveTextContent("Test Title");
});

it("does not render title when not provided", async () => {
    await act(async () => {
        render(<MinimalLayout />);
    });

    expect(screen.queryByTestId("typography-h5")).not.toBeInTheDocument();
});

it("renders children", async () => {
    await act(async () => {
        render(
            <MinimalLayout>
                <div data-testid="child">Content</div>
            </MinimalLayout>,
        );
    });

    expect(screen.getByTestId("child")).toBeInTheDocument();
});

it("sets the document title", async () => {
    await act(async () => {
        render(<MinimalLayout />);
    });

    expect(document.title).toContain("Login");
});
