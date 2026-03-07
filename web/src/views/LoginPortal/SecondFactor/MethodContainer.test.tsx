import { render, screen } from "@testing-library/react";

import MethodContainer, { State } from "@views/LoginPortal/SecondFactor/MethodContainer";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("tss-react/mui", () => ({
    makeStyles: () => () => () => ({
        classes: { container: "", containerFlex: "", containerMethod: "", info: "", infoTypography: "" },
        cx: (...args: any[]) => args.filter(Boolean).join(" "),
    }),
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

it("renders register link when onRegisterClick is provided", () => {
    render(
        <MethodContainer
            id="test"
            title="TOTP"
            duoSelfEnrollment={false}
            registered={false}
            explanation=""
            state={State.METHOD}
            onRegisterClick={vi.fn()}
        >
            <div />
        </MethodContainer>,
    );
    expect(screen.getByText("Register device")).toBeInTheDocument();
});

it("renders manage devices link when registered and onRegisterClick is provided", () => {
    render(
        <MethodContainer
            id="test"
            title="TOTP"
            duoSelfEnrollment={false}
            registered={true}
            explanation=""
            state={State.METHOD}
            onRegisterClick={vi.fn()}
        >
            <div />
        </MethodContainer>,
    );
    expect(screen.getByText("Manage devices")).toBeInTheDocument();
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
