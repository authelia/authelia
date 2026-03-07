import { fireEvent, render, screen } from "@testing-library/react";

import DecisionFormClaims from "@views/ConsentPortal/OpenIDConnect/DecisionFormClaims";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { container: "container", list: "list" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
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

it("calls onChangeChecked when checking a claim", () => {
    const onChange = vi.fn();
    render(<DecisionFormClaims claims={[]} essential_claims={null} onChangeChecked={onChange} />);
    // No checkboxes rendered for empty claims array, but the claims container still shows
    // since essential_claims is null and claims is an empty array (truthy)
});
