import { act, fireEvent, render, screen } from "@testing-library/react";

import DecisionFormPreConfiguration from "@views/ConsentPortal/OpenIDConnect/DecisionFormPreConfiguration";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { preConfigure: "preConfigure" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

it("renders checkbox when pre_configuration is true", () => {
    render(<DecisionFormPreConfiguration pre_configuration={true} onChangePreConfiguration={vi.fn()} />);
    expect(screen.getByText("Remember Consent")).toBeInTheDocument();
    expect(screen.getByRole("checkbox")).not.toBeChecked();
});

it("does not render checkbox when pre_configuration is false", () => {
    render(<DecisionFormPreConfiguration pre_configuration={false} onChangePreConfiguration={vi.fn()} />);
    expect(screen.queryByRole("checkbox")).not.toBeInTheDocument();
});

it("calls onChangePreConfiguration when checkbox is toggled", async () => {
    const onChange = vi.fn();
    render(<DecisionFormPreConfiguration pre_configuration={true} onChangePreConfiguration={onChange} />);

    await act(async () => {
        fireEvent.click(screen.getByRole("checkbox"));
    });

    expect(onChange).toHaveBeenCalledWith(true);
});
