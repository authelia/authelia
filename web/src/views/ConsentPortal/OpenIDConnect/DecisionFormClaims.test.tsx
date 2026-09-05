import { fireEvent, render, screen } from "@testing-library/react";

import DecisionFormClaims from "@views/ConsentPortal/OpenIDConnect/DecisionFormClaims";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@services/ConsentOpenIDConnect", () => ({
    formatClaim: (translated: string, claim: string) => translated || claim,
}));

it("renders nothing when no claims or essential claims", () => {
    const { container } = render(
        <DecisionFormClaims claims={null} essential_claims={null} onChangeChecked={vi.fn()} />,
    );

    expect(container).toBeEmptyDOMElement();
});

it("renders nothing when both claim lists are empty", () => {
    const { container } = render(<DecisionFormClaims claims={[]} essential_claims={[]} onChangeChecked={vi.fn()} />);

    expect(container).toBeEmptyDOMElement();
});

it("renders the section heading", () => {
    render(<DecisionFormClaims claims={["name"]} essential_claims={null} onChangeChecked={vi.fn()} />);

    expect(screen.getByRole("heading", { name: "Information Shared" })).toBeInTheDocument();
});

it("renders essential claims as disabled checkboxes", () => {
    render(<DecisionFormClaims claims={null} essential_claims={["sub", "email"]} onChangeChecked={vi.fn()} />);

    const checkboxes = screen.getAllByRole("checkbox");

    expect(checkboxes).toHaveLength(2);
    checkboxes.forEach((cb) => expect(cb).toBeDisabled());
});

it("marks essential claims as required", () => {
    render(<DecisionFormClaims claims={["name"]} essential_claims={["sub"]} onChangeChecked={vi.fn()} />);

    expect(screen.getAllByText("Required")).toHaveLength(1);
});

it("assigns a predictable id to each claim item", () => {
    render(<DecisionFormClaims claims={["name"]} essential_claims={["sub"]} onChangeChecked={vi.fn()} />);

    expect(document.getElementById("claim-sub-essential")).toBeInTheDocument();
    expect(document.getElementById("claim-name")).toBeInTheDocument();
});

it("renders optional claims as checkable checkboxes", () => {
    render(<DecisionFormClaims claims={["name", "picture"]} essential_claims={null} onChangeChecked={vi.fn()} />);

    const checkboxes = screen.getAllByRole("checkbox");

    expect(checkboxes).toHaveLength(2);
    checkboxes.forEach((cb) => expect(cb).toBeEnabled());
});

it("calls onChangeChecked when unchecking a claim", () => {
    const onChange = vi.fn();

    render(<DecisionFormClaims claims={["name", "picture"]} essential_claims={null} onChangeChecked={onChange} />);
    fireEvent.click(screen.getAllByRole("checkbox")[0]);

    expect(onChange).toHaveBeenCalledWith(["picture"]);
});

it("calls onChangeChecked when rechecking a claim", () => {
    const onChange = vi.fn();

    const { rerender } = render(
        <DecisionFormClaims claims={["name", "picture"]} essential_claims={null} onChangeChecked={onChange} />,
    );

    rerender(<DecisionFormClaims claims={["picture"]} essential_claims={null} onChangeChecked={onChange} />);
    fireEvent.click(screen.getAllByRole("checkbox")[0]);

    expect(onChange).toHaveBeenCalledWith(["picture", "name"]);
});

it("renders both essential and optional claims together", () => {
    render(<DecisionFormClaims claims={["name"]} essential_claims={["sub"]} onChangeChecked={vi.fn()} />);

    const checkboxes = screen.getAllByRole("checkbox");

    expect(checkboxes).toHaveLength(2);
    expect(checkboxes[0]).toBeDisabled();
    expect(checkboxes[1]).toBeEnabled();
});

it("shows the raw claim alongside its description", () => {
    render(<DecisionFormClaims claims={["email"]} essential_claims={null} onChangeChecked={vi.fn()} />);

    expect(screen.getByText("email")).toBeInTheDocument();
});

it("shows the raw claim alongside the required marker for essential claims", () => {
    render(<DecisionFormClaims claims={null} essential_claims={["sub"]} onChangeChecked={vi.fn()} />);

    expect(screen.getByText("sub")).toBeInTheDocument();
    expect(screen.getByText("Required")).toBeInTheDocument();
});
