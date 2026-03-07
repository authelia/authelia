import { render, screen } from "@testing-library/react";

import AuthenticatedView from "@views/LoginPortal/AuthenticatedView/AuthenticatedView";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { mainContainer: "mainContainer" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@components/LogoutButton", () => ({
    default: () => <button data-testid="logout-button">Logout</button>,
}));

vi.mock("@layouts/MinimalLayout", () => ({
    default: (props: any) => (
        <div data-testid="minimal-layout" data-title={props.title}>
            {props.children}
        </div>
    ),
}));

vi.mock("@views/LoginPortal/Authenticated", () => ({
    default: () => <div data-testid="authenticated" />,
}));

it("renders with user display name in title", () => {
    render(<AuthenticatedView userInfo={{ display_name: "John", emails: [], groups: [], method: "totp" } as any} />);
    expect(screen.getByTestId("minimal-layout")).toHaveAttribute("data-title", "Hi John");
});

it("renders logout button and authenticated component", () => {
    render(<AuthenticatedView userInfo={{ display_name: "Jane", emails: [], groups: [], method: "totp" } as any} />);
    expect(screen.getByTestId("logout-button")).toBeInTheDocument();
    expect(screen.getByTestId("authenticated")).toBeInTheDocument();
});
