import { act, fireEvent, render, screen } from "@testing-library/react";

import DecisionFormPreConfiguration from "@views/ConsentPortal/OpenIDConnect/DecisionFormPreConfiguration";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

it("renders the switch when pre_configuration is true", () => {
    render(<DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={vi.fn()} />);

    expect(screen.getByText("Remember Consent")).toBeInTheDocument();
    expect(screen.getByRole("switch")).not.toBeChecked();
});

it("does not render the switch when pre_configuration is false", () => {
    render(
        <DecisionFormPreConfiguration pre_configuration={false} checked={false} onChangePreConfiguration={vi.fn()} />,
    );

    expect(screen.queryByRole("switch")).not.toBeInTheDocument();
});

it("reflects the checked value it is given", () => {
    render(<DecisionFormPreConfiguration pre_configuration checked onChangePreConfiguration={vi.fn()} />);

    expect(screen.getByRole("switch")).toBeChecked();
});

it("calls onChangePreConfiguration when the switch is toggled", async () => {
    const onChange = vi.fn();

    render(<DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={onChange} />);

    await act(async () => {
        fireEvent.click(screen.getByRole("switch"));
    });

    expect(onChange).toHaveBeenCalledWith(true);
});

it("reports the change back to false when unchecked", async () => {
    const onChange = vi.fn();

    render(<DecisionFormPreConfiguration pre_configuration checked onChangePreConfiguration={onChange} />);

    await act(async () => {
        fireEvent.click(screen.getByRole("switch"));
    });

    expect(onChange).toHaveBeenCalledWith(false);
});

it("does not report a change on mount", () => {
    const onChange = vi.fn();

    render(<DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={onChange} />);

    expect(onChange).not.toHaveBeenCalled();
});

it("describes what remembering the consent does", () => {
    render(<DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={vi.fn()} />);

    expect(screen.getByText("This saves this consent as a pre-configured consent for future use")).toBeInTheDocument();
});

it("labels the switch", () => {
    render(<DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={vi.fn()} />);

    expect(screen.getByRole("switch")).toHaveAccessibleName("Remember Consent");
});

it("renders as a muted item", () => {
    const { container } = render(
        <DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={vi.fn()} />,
    );

    const item = container.querySelector("[data-slot='item']");

    expect(item).toBeInTheDocument();
    expect(item).toHaveAttribute("data-variant", "muted");
});

it("places the switch in the item actions", () => {
    const { container } = render(
        <DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={vi.fn()} />,
    );

    expect(container.querySelector("[data-slot='item-actions']")).toContainElement(screen.getByRole("switch"));
});

it("renders the explanation as the item description", () => {
    const { container } = render(
        <DecisionFormPreConfiguration pre_configuration checked={false} onChangePreConfiguration={vi.fn()} />,
    );

    expect(container.querySelector("[data-slot='item-description']")).toHaveTextContent(
        "This saves this consent as a pre-configured consent for future use",
    );
});
