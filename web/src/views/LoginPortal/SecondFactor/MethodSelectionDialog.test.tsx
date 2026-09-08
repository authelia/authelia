import { fireEvent, render, screen, waitFor } from "@testing-library/react";

import { SecondFactorMethod } from "@models/Methods";
import MethodSelectionDialog from "@views/LoginPortal/SecondFactor/MethodSelectionDialog";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@components/FingerTouchIcon", () => ({
    default: () => <div data-testid="finger-touch-icon" />,
}));

vi.mock("@components/PushNotificationIcon", () => ({
    default: () => <div data-testid="push-icon" />,
}));

vi.mock("@components/TimerIcon", () => ({
    default: () => <div data-testid="timer-icon" />,
}));

it("renders TOTP option when available", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.getByText("Time-based One-Time Password")).toBeInTheDocument();
});

it("renders WebAuthn option when available and supported", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.WebAuthn])}
            webauthn={true}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.getByText("Security Key - WebAuthn")).toBeInTheDocument();
});

it("does not render WebAuthn option when not supported", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.WebAuthn])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.queryByText("Security Key - WebAuthn")).not.toBeInTheDocument();
});

it("renders Push Notification option when available", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.MobilePush])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );
    expect(screen.getByText("Push Notification")).toBeInTheDocument();
});

it("calls onClick with the correct method", () => {
    const onClick = vi.fn();
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={onClick}
        />,
    );
    fireEvent.click(screen.getByText("Time-based One-Time Password"));
    expect(onClick).toHaveBeenCalledWith(SecondFactorMethod.TOTP);
});

it("calls onClose when Close button is clicked", () => {
    const onClose = vi.fn();
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={onClose}
            onClick={vi.fn()}
        />,
    );
    const closeButtons = screen.getAllByRole("button", { name: "Close" });
    const footerCloseButton = closeButtons.find((btn) => btn.textContent === "Close")!;
    fireEvent.click(footerCloseButton);
    expect(onClose).toHaveBeenCalledOnce();
});

it("calls onClick with WebAuthn", () => {
    const onClick = vi.fn();
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.WebAuthn])}
            webauthn={true}
            onClose={vi.fn()}
            onClick={onClick}
        />,
    );
    fireEvent.click(screen.getByText("Security Key - WebAuthn"));
    expect(onClick).toHaveBeenCalledWith(SecondFactorMethod.WebAuthn);
});

it("calls onClick with MobilePush", () => {
    const onClick = vi.fn();
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.MobilePush])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={onClick}
        />,
    );
    fireEvent.click(screen.getByText("Push Notification"));
    expect(onClick).toHaveBeenCalledWith(SecondFactorMethod.MobilePush);
});

it("renders every available method together", () => {
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP, SecondFactorMethod.WebAuthn, SecondFactorMethod.MobilePush])}
            webauthn={true}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );

    expect(document.getElementById("one-time-password-option")).toBeInTheDocument();
    expect(document.getElementById("webauthn-option")).toBeInTheDocument();
    expect(document.getElementById("push-notification-option")).toBeInTheDocument();
    expect(screen.getByTestId("timer-icon")).toBeInTheDocument();
    expect(screen.getByTestId("finger-touch-icon")).toBeInTheDocument();
    expect(screen.getByTestId("push-icon")).toBeInTheDocument();
});

it("renders no options when none are available", () => {
    render(
        <MethodSelectionDialog open={true} methods={new Set()} webauthn={true} onClose={vi.fn()} onClick={vi.fn()} />,
    );

    expect(document.getElementById("methods-dialog")?.children).toHaveLength(0);
});

it("does not render content when closed", () => {
    render(
        <MethodSelectionDialog
            open={false}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={vi.fn()}
            onClick={vi.fn()}
        />,
    );

    expect(screen.queryByText("Time-based One-Time Password")).not.toBeInTheDocument();
});

it("closes when Escape is pressed", async () => {
    const onClose = vi.fn();
    render(
        <MethodSelectionDialog
            open={true}
            methods={new Set([SecondFactorMethod.TOTP])}
            webauthn={false}
            onClose={onClose}
            onClick={vi.fn()}
        />,
    );

    fireEvent.keyDown(document.activeElement ?? document.body, { key: "Escape" });

    await waitFor(() => expect(onClose).toHaveBeenCalled());
});
