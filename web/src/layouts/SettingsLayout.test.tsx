import { act, fireEvent, render, screen } from "@testing-library/react";

import SettingsLayout from "@layouts/SettingsLayout";

const mockNavigate = vi.fn();

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@constants/constants", () => ({
    EncodedName: [81, 88, 86, 48, 97, 71, 86, 115, 97, 87, 69, 61],
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
    SecuritySubRoute: "/security",
    SettingsRoute: "/settings",
    SettingsTwoFactorAuthenticationSubRoute: "/two-factor-authentication",
}));

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: () => mockNavigate,
}));

beforeEach(() => {
    mockNavigate.mockReset();
});

it("renders with settings title and menu button", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    expect(screen.getByLabelText("open drawer")).toBeInTheDocument();
    expect(screen.getAllByText("Settings").length).toBeGreaterThanOrEqual(1);
});

it("renders children", async () => {
    await act(async () => {
        render(
            <SettingsLayout>
                <div data-testid="child">Content</div>
            </SettingsLayout>,
        );
    });

    expect(screen.getByTestId("child")).toBeInTheDocument();
});

it("renders navigation items in drawer", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    expect(screen.getByText("Overview")).toBeInTheDocument();
    expect(screen.getByText("Security")).toBeInTheDocument();
    expect(screen.getByText("Two-Factor Authentication")).toBeInTheDocument();
    expect(screen.getByText("Close")).toBeInTheDocument();
});

it("sets the document title", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    expect(document.title).toContain("Settings");
});

it("opens drawer when menu button is clicked", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    expect(screen.getByRole("presentation")).toBeInTheDocument();
});

it("navigates when a nav item is clicked", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Security"));
    });

    expect(mockNavigate).toHaveBeenCalledWith("/settings/security");
});

it("does not navigate when the selected nav item is clicked", async () => {
    Object.defineProperty(globalThis, "location", {
        configurable: true,
        value: { pathname: "/settings" },
        writable: true,
    });

    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByText("Overview"));
    });

    expect(mockNavigate).not.toHaveBeenCalled();

    Object.defineProperty(globalThis, "location", {
        configurable: true,
        value: { pathname: "/" },
        writable: true,
    });
});

it("does not close drawer on Tab keydown event", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    const drawerContent = screen.getByText("Overview").closest("[role='presentation']")!;

    await act(async () => {
        const event = new KeyboardEvent("keydown", { bubbles: true, key: "Tab" });
        drawerContent.dispatchEvent(event);
    });

    expect(screen.getByRole("presentation")).toBeInTheDocument();
});

it("does not close drawer on Shift keydown event", async () => {
    await act(async () => {
        render(<SettingsLayout />);
    });

    await act(async () => {
        fireEvent.click(screen.getByLabelText("open drawer"));
    });

    const drawerContent = screen.getByText("Overview").closest("[role='presentation']")!;

    await act(async () => {
        const event = new KeyboardEvent("keydown", { bubbles: true, key: "Shift" });
        drawerContent.dispatchEvent(event);
    });

    expect(screen.getByRole("presentation")).toBeInTheDocument();
});
