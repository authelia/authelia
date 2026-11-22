import { useState } from "react";

import { fireEvent, render, screen } from "@testing-library/react";

import { Combobox } from "@components/UI/Combobox";

const options = ["Alpha", "Beta", "Gamma"];

function ControlledCombobox({
    freeSolo,
    initial = [],
    onChange,
}: {
    freeSolo?: boolean;
    initial?: string[];
    onChange: (value: string[]) => void;
}) {
    const [value, setValue] = useState<string[]>(initial);
    return (
        <Combobox
            freeSolo={freeSolo}
            id="my-combobox"
            onChange={(next) => {
                setValue(next);
                onChange(next);
            }}
            options={options}
            value={value}
        />
    );
}

it("renders with the combobox data-slot and merges a passed className", () => {
    const { container } = render(
        <Combobox className="custom-combobox" id="cb" onChange={vi.fn()} options={options} value={[]} />,
    );
    const el = container.querySelector('[data-slot="combobox"]');
    expect(el).toBeInTheDocument();
    expect(el).toHaveClass("custom-combobox");
});

it("exposes the input with the given id and combobox role", () => {
    render(<Combobox id="my-combobox" onChange={vi.fn()} options={options} value={[]} />);
    const input = document.getElementById("my-combobox");
    expect(input).not.toBeNull();
    expect(input).toHaveAttribute("role", "combobox");
});

it("renders selected values as badges", () => {
    render(<Combobox id="cb" onChange={vi.fn()} options={options} value={["Alpha"]} />);
    const badge = screen.getByText("Alpha").closest('[data-slot="combobox-badge"]');
    expect(badge).toBeInTheDocument();
});

it("selecting an option calls onChange with the updated array", async () => {
    const onChange = vi.fn();
    render(<ControlledCombobox onChange={onChange} />);

    const input = document.getElementById("my-combobox")!;
    fireEvent.focus(input);
    fireEvent.pointerDown(input);
    fireEvent.mouseDown(input);
    fireEvent.click(input);
    fireEvent.change(input, { target: { value: "Beta" } });

    const option = await screen.findByText("Beta");
    fireEvent.click(option);

    expect(onChange).toHaveBeenCalledWith(["Beta"]);
});

it("removing a badge removes that value", () => {
    const onChange = vi.fn();
    render(<ControlledCombobox initial={["Alpha", "Beta"]} onChange={onChange} />);

    fireEvent.click(screen.getByLabelText("Remove Alpha"));

    expect(onChange).toHaveBeenCalledWith(["Beta"]);
});

it("freeSolo: pressing Enter on typed text adds it to the array", () => {
    const onChange = vi.fn();
    render(<ControlledCombobox freeSolo onChange={onChange} />);

    const input = document.getElementById("my-combobox")!;
    fireEvent.change(input, { target: { value: "Custom Value" } });
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onChange).toHaveBeenCalledWith(["Custom Value"]);
});

it("freeSolo: pressing comma on typed text adds it to the array", () => {
    const onChange = vi.fn();
    render(<ControlledCombobox freeSolo onChange={onChange} />);

    const input = document.getElementById("my-combobox")!;
    fireEvent.change(input, { target: { value: "Another" } });
    fireEvent.keyDown(input, { key: "," });

    expect(onChange).toHaveBeenCalledWith(["Another"]);
});

it("without freeSolo, pressing Enter on typed text does not add it", () => {
    const onChange = vi.fn();
    render(<ControlledCombobox onChange={onChange} />);

    const input = document.getElementById("my-combobox")!;
    fireEvent.change(input, { target: { value: "Not An Option" } });
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onChange).not.toHaveBeenCalled();
});
