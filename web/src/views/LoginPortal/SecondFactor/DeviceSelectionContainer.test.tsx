import { fireEvent, render, screen } from "@testing-library/react";

import DeviceSelectionContainer from "@views/LoginPortal/SecondFactor/DeviceSelectionContainer";

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { buttonRoot: "", icon: "", item: "" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
}));

vi.mock("@components/PushNotificationIcon", () => ({
    default: () => <div data-testid="push-icon" />,
}));

const devices = [
    { id: "dev1", methods: ["push"], name: "Phone" },
    { id: "dev2", methods: ["push", "sms"], name: "Tablet" },
];

it("renders device list", () => {
    render(<DeviceSelectionContainer devices={devices} onBack={vi.fn()} onSelect={vi.fn()} />);
    expect(screen.getByText("Phone")).toBeInTheDocument();
    expect(screen.getByText("Tablet")).toBeInTheDocument();
});

it("calls onSelect directly when device has single method", () => {
    const onSelect = vi.fn();
    render(<DeviceSelectionContainer devices={devices} onBack={vi.fn()} onSelect={onSelect} />);
    fireEvent.click(screen.getByText("Phone"));
    expect(onSelect).toHaveBeenCalledWith({ id: "dev1", method: "push" });
});

it("shows method selection when device has multiple methods", () => {
    render(<DeviceSelectionContainer devices={devices} onBack={vi.fn()} onSelect={vi.fn()} />);
    fireEvent.click(screen.getByText("Tablet"));
    expect(screen.getByText("push")).toBeInTheDocument();
    expect(screen.getByText("sms")).toBeInTheDocument();
});

it("calls onSelect when a method is chosen from multi-method device", () => {
    const onSelect = vi.fn();
    render(<DeviceSelectionContainer devices={devices} onBack={vi.fn()} onSelect={onSelect} />);
    fireEvent.click(screen.getByText("Tablet"));
    fireEvent.click(screen.getByText("sms"));
    expect(onSelect).toHaveBeenCalledWith({ id: "dev2", method: "sms" });
});

it("calls onBack when back button is clicked", () => {
    const onBack = vi.fn();
    render(<DeviceSelectionContainer devices={devices} onBack={onBack} onSelect={vi.fn()} />);
    fireEvent.click(screen.getByText("back"));
    expect(onBack).toHaveBeenCalledOnce();
});
