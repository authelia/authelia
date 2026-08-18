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
    expect(container.querySelectorAll("label")).toHaveLength(0);
});

it("renders essential claims as disabled checkboxes", () => {
    render(<DecisionFormClaims claims={null} essential_claims={["sub", "email"]} onChangeChecked={vi.fn()} />);
    const checkboxes = screen.getAllByRole("checkbox");
    expect(checkboxes).toHaveLength(2);
    checkboxes.forEach((cb) => expect(cb).toBeDisabled());
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

it("renders both essential and optional claims together", () => {
    render(<DecisionFormClaims claims={["name"]} essential_claims={["sub"]} onChangeChecked={vi.fn()} />);
    const checkboxes = screen.getAllByRole("checkbox");
    expect(checkboxes).toHaveLength(2);
    expect(checkboxes[0]).toBeDisabled();
    expect(checkboxes[1]).toBeEnabled();
});
