import { render, screen } from "@testing-library/react";

import OTPDial, { State } from "@views/LoginPortal/SecondFactor/OTPDial";

vi.mock("react18-input-otp", () => ({
    default: (props: any) => (
        <div data-testid="otp-input" data-disabled={props.isDisabled} data-num-inputs={props.numInputs} />
    ),
}));

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

it("renders OTP input with correct digit count", () => {
    render(<OTPDial passcode="" state={State.Idle} digits={6} period={30} onChange={vi.fn()} />);
    expect(screen.getByTestId("otp-input")).toHaveAttribute("data-num-inputs", "6");
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
    expect(screen.getByTestId("otp-input")).toHaveAttribute("data-disabled", "true");
});

it("disables input during rate-limited state", () => {
    render(<OTPDial passcode="" state={State.RateLimited} digits={6} period={30} onChange={vi.fn()} />);
    expect(screen.getByTestId("otp-input")).toHaveAttribute("data-disabled", "true");
});
