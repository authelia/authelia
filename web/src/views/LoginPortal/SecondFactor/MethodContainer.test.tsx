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

it("renders the device selection link when a handler is provided and the user is registered", () => {
    const onSelectClick = vi.fn();

    render(
        <MethodContainer
            id="test"
            title="Security Key"
            duoSelfEnrollment={false}
            registered={true}
            explanation="Touch it"
            state={State.METHOD}
            onSelectClick={onSelectClick}
        >
            <div />
        </MethodContainer>,
    );

    fireEvent.click(screen.getByText("Select a Device"));

    expect(onSelectClick).toHaveBeenCalledTimes(1);
});

it("omits the device selection link when the user is not registered", () => {
    render(
        <MethodContainer
            id="test"
            title="Security Key"
            duoSelfEnrollment={false}
            registered={false}
            explanation="Touch it"
            state={State.METHOD}
            onSelectClick={vi.fn()}
        >
            <div />
        </MethodContainer>,
    );

    expect(screen.queryByText("Select a Device")).not.toBeInTheDocument();
});

it("omits the device selection link without a handler", () => {
    render(
        <MethodContainer
            id="test"
            title="Security Key"
            duoSelfEnrollment={false}
            registered={true}
            explanation="Touch it"
            state={State.METHOD}
        >
            <div />
        </MethodContainer>,
    );

    expect(screen.queryByText("Select a Device")).not.toBeInTheDocument();
});

it("omits the register link for push notifications without self enrollment", () => {
    render(
        <MethodContainer
            id="test"
            title="Push Notification"
            duoSelfEnrollment={false}
            registered={false}
            explanation=""
            state={State.NOT_REGISTERED}
            onRegisterClick={vi.fn()}
        >
            <div />
        </MethodContainer>,
    );

    expect(document.getElementById("register-link")).toBeNull();
});

it("renders the register link for push notifications with self enrollment", () => {
    const onRegisterClick = vi.fn();

    render(
        <MethodContainer
            id="test"
            title="Push Notification"
            duoSelfEnrollment={true}
            registered={false}
            explanation=""
            state={State.NOT_REGISTERED}
            onRegisterClick={onRegisterClick}
        >
            <div />
        </MethodContainer>,
    );

    fireEvent.click(document.getElementById("register-link") as HTMLElement);

    expect(onRegisterClick).toHaveBeenCalledTimes(1);
});

it("offers no manage devices message for a registered push notification method", () => {
    render(
        <MethodContainer
            id="test"
            title="Push Notification"
            duoSelfEnrollment={true}
            registered={true}
            explanation=""
            state={State.METHOD}
            onRegisterClick={vi.fn()}
        >
            <div />
        </MethodContainer>,
    );

    expect(document.getElementById("register-link")).toBeEmptyDOMElement();
    expect(screen.queryByText("Manage devices")).not.toBeInTheDocument();
});

it("renders the self enrollment prompt for push notifications", () => {
    render(
        <MethodContainer
            id="test"
            title="Push Notification"
            duoSelfEnrollment={true}
            registered={false}
            explanation=""
            state={State.NOT_REGISTERED}
        >
            <div />
        </MethodContainer>,
    );

    expect(screen.getByText("Register your first device by clicking on the link below")).toBeInTheDocument();
});
