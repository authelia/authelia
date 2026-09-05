import { render, screen } from "@testing-library/react";
import i18n from "i18next";

import DecisionFormClaims from "@views/ConsentPortal/OpenIDConnect/DecisionFormClaims";
import DecisionFormScopes from "@views/ConsentPortal/OpenIDConnect/DecisionFormScopes";

beforeAll(async () => {
    i18n.addResourceBundle("en", "consent", {
        claims: { email: "E-mail Address" },
        "Information Shared": "Information Shared",
        "Requested Permissions": "Requested Permissions",
        scopes: { openid: "Use OpenID to verify your identity" },
    });

    await i18n.changeLanguage("en");
});

it("renders a known scope with its description", () => {
    render(<DecisionFormScopes scopes={["openid"]} />);

    expect(screen.getByText("Use OpenID to verify your identity")).toBeInTheDocument();
});

it("renders a dotted custom scope verbatim", () => {
    render(<DecisionFormScopes scopes={["metrics.read"]} />);

    expect(screen.getByText("metrics.read")).toBeInTheDocument();
});

it("renders a colon delimited custom scope verbatim", () => {
    render(<DecisionFormScopes scopes={["api:read"]} />);

    expect(screen.getByText("api:read")).toBeInTheDocument();
});

it("does not truncate a colon delimited custom scope to its suffix", () => {
    render(<DecisionFormScopes scopes={["read:user"]} />);

    expect(screen.queryByText("user")).not.toBeInTheDocument();
    expect(screen.getByText("read:user")).toBeInTheDocument();
});

it("renders a known claim with its description", () => {
    render(
        <DecisionFormClaims claims={["email"]} checked={["email"]} essential_claims={null} onChangeChecked={vi.fn()} />,
    );

    expect(screen.getByText("E-mail Address")).toBeInTheDocument();
});

it("does not truncate a colon delimited custom claim to its suffix", () => {
    render(
        <DecisionFormClaims
            claims={["urn:example:role"]}
            checked={["urn:example:role"]}
            essential_claims={null}
            onChangeChecked={vi.fn()}
        />,
    );

    expect(screen.queryByText("role")).not.toBeInTheDocument();
    expect(screen.getByText("Urn:example:role")).toBeInTheDocument();
});

it("shows the raw scope beside a described scope", () => {
    render(<DecisionFormScopes scopes={["openid"]} />);

    expect(screen.getByText("Use OpenID to verify your identity")).toBeInTheDocument();
    expect(screen.getByText("openid")).toBeInTheDocument();
});

it("does not repeat a custom scope that is already its own label", () => {
    render(<DecisionFormScopes scopes={["dashboards:write"]} />);

    expect(screen.getAllByText("dashboards:write")).toHaveLength(1);
});

it("does not repeat a custom claim that differs from its label only by case", () => {
    render(
        <DecisionFormClaims
            claims={["urn:example:role"]}
            checked={["urn:example:role"]}
            essential_claims={null}
            onChangeChecked={vi.fn()}
        />,
    );

    expect(screen.getAllByText(/urn:example:role/i)).toHaveLength(1);
});

it("shows the raw claim beside a described claim", () => {
    render(
        <DecisionFormClaims claims={["email"]} checked={["email"]} essential_claims={null} onChangeChecked={vi.fn()} />,
    );

    expect(screen.getByText("E-mail Address")).toBeInTheDocument();
    expect(screen.getByText("email")).toBeInTheDocument();
});
