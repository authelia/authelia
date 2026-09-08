import { fireEvent, render, screen } from "@testing-library/react";

import DecisionFormRequest from "@views/ConsentPortal/OpenIDConnect/DecisionFormRequest";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@services/ConsentOpenIDConnect", () => ({
    formatClaim: (translated: string, claim: string) => translated || claim,
    formatScope: (translated: string, scope: string) => translated || scope,
}));

const response = (overrides: Partial<any> = {}) =>
    ({
        audience: ["https://api.example.com"],
        claims: ["name"],
        client_description: "Test Client",
        client_id: "test-client",
        essential_claims: ["sub"],
        pre_configuration: false,
        require_login: false,
        resource: ["https://files.example.com"],
        scopes: ["openid", "profile"],
        ...overrides,
    }) as any;

it("renders the client header", () => {
    render(<DecisionFormRequest response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    expect(screen.getByTestId("openid-consent-client-name")).toHaveTextContent("Test Client");
});

it("renders every requested section", () => {
    render(<DecisionFormRequest response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    expect(document.getElementById("openid-consent-scopes")).toBeInTheDocument();
    expect(document.getElementById("openid-consent-claims")).toBeInTheDocument();
    expect(document.getElementById("openid-consent-audience")).toBeInTheDocument();
    expect(document.getElementById("openid-consent-resource")).toBeInTheDocument();
});

it("does not render a disclosure when it is not collapsible", () => {
    render(<DecisionFormRequest response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    expect(screen.queryByRole("button", { name: "Show request details" })).not.toBeInTheDocument();
});

it("renders a disclosure when it is collapsible", () => {
    render(<DecisionFormRequest collapsible response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    expect(screen.getByRole("button", { name: "Show request details" })).toBeInTheDocument();
});

it("keeps the client header visible when it is collapsible", () => {
    render(<DecisionFormRequest collapsible response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    expect(screen.getByTestId("openid-consent-client-name")).toHaveTextContent("Test Client");
});

it("hides the sections until the disclosure is opened", () => {
    render(<DecisionFormRequest collapsible response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    expect(document.getElementById("openid-consent-scopes")).not.toBeInTheDocument();
});

it("reveals the sections when the disclosure is opened", () => {
    render(<DecisionFormRequest collapsible response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "Show request details" }));

    expect(document.getElementById("openid-consent-scopes")).toBeInTheDocument();
    expect(document.getElementById("openid-consent-audience")).toBeInTheDocument();
});

it("offers to hide the sections once the disclosure is open", () => {
    render(<DecisionFormRequest collapsible response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "Show request details" }));

    expect(screen.getByRole("button", { name: "Hide request details" })).toBeInTheDocument();
});

it("hides the sections again when the disclosure is closed", () => {
    render(<DecisionFormRequest collapsible response={response()} claims={["name"]} onChangeClaims={vi.fn()} />);

    fireEvent.click(screen.getByRole("button", { name: "Show request details" }));
    fireEvent.click(screen.getByRole("button", { name: "Hide request details" }));

    expect(document.getElementById("openid-consent-scopes")).not.toBeInTheDocument();
});

it("does not render a disclosure when there is nothing to disclose", () => {
    render(
        <DecisionFormRequest
            collapsible
            response={response({ audience: [], claims: null, essential_claims: null, resource: null, scopes: [] })}
            claims={[]}
            onChangeClaims={vi.fn()}
        />,
    );

    expect(screen.queryByRole("button", { name: "Show request details" })).not.toBeInTheDocument();
});

it("renders a disclosure when every optional claim is deselected", () => {
    render(
        <DecisionFormRequest
            collapsible
            response={response({ audience: [], essential_claims: null, resource: null, scopes: [] })}
            claims={[]}
            onChangeClaims={vi.fn()}
        />,
    );

    expect(screen.getByRole("button", { name: "Show request details" })).toBeInTheDocument();
});

it("keeps deselected claims listed after the request details are reopened", () => {
    const { rerender } = render(
        <DecisionFormRequest
            response={response({ claims: ["name", "picture"] })}
            claims={["name", "picture"]}
            onChangeClaims={vi.fn()}
        />,
    );

    rerender(
        <DecisionFormRequest
            collapsible
            response={response({ claims: ["name", "picture"] })}
            claims={["picture"]}
            onChangeClaims={vi.fn()}
        />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Show request details" }));

    expect(document.getElementById("claim-name")).toBeInTheDocument();
    expect(document.getElementById("claim-picture")).toBeInTheDocument();
});

it("reports claim changes", () => {
    const onChangeClaims = vi.fn();

    render(<DecisionFormRequest response={response()} claims={["name"]} onChangeClaims={onChangeClaims} />);

    fireEvent.click(document.getElementById("claim-name") as HTMLElement);

    expect(onChangeClaims).toHaveBeenCalledWith([]);
});
