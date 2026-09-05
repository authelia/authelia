import { render, screen } from "@testing-library/react";

import DecisionFormScopes from "@views/ConsentPortal/OpenIDConnect/DecisionFormScopes";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@components/OpenIDConnect", () => ({
    ScopeAvatar: () => <span data-testid="scope-avatar" />,
}));

vi.mock("@services/ConsentOpenIDConnect", () => ({
    formatScope: (translated: string, scope: string) => translated || scope,
}));

it("renders scope items", () => {
    render(<DecisionFormScopes scopes={["openid", "profile", "email"]} />);

    expect(screen.getByText("scopes.openid")).toBeInTheDocument();
    expect(screen.getByText("scopes.profile")).toBeInTheDocument();
    expect(screen.getByText("scopes.email")).toBeInTheDocument();
});

it("renders the section heading", () => {
    render(<DecisionFormScopes scopes={["openid"]} />);

    expect(screen.getByRole("heading", { name: "Requested Permissions" })).toBeInTheDocument();
});

it("assigns a predictable id to each scope item", () => {
    render(<DecisionFormScopes scopes={["openid"]} />);

    expect(screen.getByRole("listitem")).toHaveAttribute("id", "scope-openid");
});

it("renders an avatar for each scope", () => {
    render(<DecisionFormScopes scopes={["openid", "profile"]} />);

    expect(screen.getAllByTestId("scope-avatar")).toHaveLength(2);
});

it("renders nothing when no scopes", () => {
    const { container } = render(<DecisionFormScopes scopes={[]} />);

    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when the scopes are null", () => {
    const { container } = render(<DecisionFormScopes scopes={null} />);

    expect(container).toBeEmptyDOMElement();
});

it("shows the raw scope alongside its description", () => {
    render(<DecisionFormScopes scopes={["openid"]} />);

    expect(screen.getByText("openid")).toBeInTheDocument();
});
