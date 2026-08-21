import { render, screen } from "@testing-library/react";

import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

vi.mock("@components/SuccessIcon", () => ({
    default: () => <div data-testid="success-icon" />,
}));

vi.mock("@components/TimerIcon", () => ({
    default: () => <div data-testid="timer-icon" />,
}));

vi.mock("@views/LoginPortal/SecondFactor/IconWithContext", () => ({
    default: (props: any) => (
        <div data-testid="icon-context">
            {props.icon}
            {props.children}
        </div>
    ),
}));

const inputs = () =>
    Array.from(document.querySelectorAll("#otp-input input")).filter(
        (el) => el.getAttribute("aria-hidden") !== "true",
    ) as HTMLInputElement[];

it("renders OTP input with correct digit count", () => {
    render(<OTPDial passcode="" state={State.Idle} digits={6} period={30} onChange={vi.fn()} />);
    expect(inputs()).toHaveLength(6);
});

it("renders timer icon in idle state", () => {
    render(<OTPDial passcode="" state={State.Idle} digits={6} period={30} onChange={vi.fn()} />);
    expect(screen.getByTestId("timer-icon")).toBeInTheDocument();
});

it("renders success icon in success state", () => {
    render(<OTPDial passcode="123456" state={State.Success} digits={6} period={30} onChange={vi.fn()} />);
    expect(screen.getByTestId("success-icon")).toBeInTheDocument();
});

it("disables input during in-progress state", () => {
    render(<OTPDial passcode="" state={State.InProgress} digits={6} period={30} onChange={vi.fn()} />);
    expect(inputs().every((el) => el.disabled)).toBe(true);
});

it("disables input during rate-limited state", () => {
    render(<OTPDial passcode="" state={State.RateLimited} digits={6} period={30} onChange={vi.fn()} />);
    expect(inputs().every((el) => el.disabled)).toBe(true);
});

it("names the field and marks the failure state for assistive technology", () => {
    const { container, unmount } = render(
        <OTPDial passcode="" state={State.Idle} digits={6} period={30} onChange={vi.fn()} />,
    );

    const labeledBy = inputs()[0].getAttribute("aria-labelledby");
    expect(labeledBy).toBeTruthy();
    expect(container.querySelector(`#${CSS.escape(labeledBy!)}`)?.textContent).toBe("Enter One-Time Password");
    expect(inputs().every((el) => el.getAttribute("aria-labelledby") === labeledBy)).toBe(true);
    expect(inputs()[0]).toHaveAttribute("aria-invalid", "false");
    unmount();

    render(<OTPDial passcode="123" state={State.Failure} digits={6} period={30} onChange={vi.fn()} />);
    expect(inputs().every((el) => el.getAttribute("aria-invalid") === "true")).toBe(true);
});
