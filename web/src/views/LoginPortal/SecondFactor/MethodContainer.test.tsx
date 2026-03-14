import { fireEvent, render, screen } from "@testing-library/react";

import MethodContainer, { State } from "@views/LoginPortal/SecondFactor/MethodContainer";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@components/InformationIcon", () => ({
    default: () => <div data-testid="info-icon" />,
}));

vi.mock("@views/LoginPortal/Authenticated", () => ({
    default: () => <div data-testid="authenticated" />,
}));

it("renders authenticated state", () => {
    render(
        <MethodContainer
            id="test"
            title="TOTP"
            duoSelfEnrollment={false}
            registered={true}
            explanation=""
            state={State.ALREADY_AUTHENTICATED}
        >
            <div />
        </MethodContainer>,
    );
    expect(screen.getByTestId("authenticated")).toBeInTheDocument();
});

it("renders not registered state", () => {
    render(
        <MethodContainer
            id="test"
            title="TOTP"
            duoSelfEnrollment={false}
            registered={false}
            explanation=""
            state={State.NOT_REGISTERED}
        >
            <div />
        </MethodContainer>,
    );
    expect(screen.getByText("Register your first device by clicking on the link below")).toBeInTheDocument();
});

it("renders method state with children and explanation", () => {
    render(
        <MethodContainer
            id="test"
            title="TOTP"
            duoSelfEnrollment={false}
            registered={true}
            explanation="Enter your code"
            state={State.METHOD}
        >
            <div data-testid="method-child" />
        </MethodContainer>,
    );
    expect(screen.getByTestId("method-child")).toBeInTheDocument();
    expect(screen.getByText("Enter your code")).toBeInTheDocument();
});

it("renders register link and calls onRegisterClick when clicked", () => {
    const onRegisterClick = vi.fn();
    render(
        <MethodContainer
            id="test"
            title="TOTP"
            duoSelfEnrollment={false}
            registered={false}
            explanation=""
            state={State.METHOD}
            onRegisterClick={onRegisterClick}
        >
            <div />
        </MethodContainer>,
    );
    const link = screen.getByText("Register device");
    expect(link).toBeInTheDocument();
    fireEvent.click(link);
    expect(onRegisterClick).toHaveBeenCalledTimes(1);
});

it("renders manage devices link and calls onRegisterClick when clicked", () => {
    const onRegisterClick = vi.fn();
    render(
        <MethodContainer
            id="test"
            title="TOTP"
            duoSelfEnrollment={false}
            registered={true}
            explanation=""
            state={State.METHOD}
            onRegisterClick={onRegisterClick}
        >
            <div />
        </MethodContainer>,
    );
    const link = screen.getByText("Manage devices");
    expect(link).toBeInTheDocument();
    fireEvent.click(link);
    expect(onRegisterClick).toHaveBeenCalledTimes(1);
});

it("renders push notification not registered state without self enrollment", () => {
    render(
        <MethodContainer
            id="test"
            title="Push Notification"
            duoSelfEnrollment={false}
            registered={false}
            explanation=""
            state={State.NOT_REGISTERED}
        >
            <div />
        </MethodContainer>,
    );
    expect(screen.getByText("Contact your administrator to register a device")).toBeInTheDocument();
});
