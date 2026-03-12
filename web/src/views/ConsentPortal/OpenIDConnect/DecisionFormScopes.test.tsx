import { render, screen } from "@testing-library/react";

import DecisionFormScopes from "@views/ConsentPortal/OpenIDConnect/DecisionFormScopes";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { scopesList: "scopesList", scopesListContainer: "scopesListContainer" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
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

it("renders empty list when no scopes", () => {
    const { container } = render(<DecisionFormScopes scopes={[]} />);
    expect(container.querySelectorAll("li")).toHaveLength(0);
});
