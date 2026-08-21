import { fireEvent, render, screen } from "@testing-library/react";

import TwoFactorAuthenticationOptionsMethodsRadioGroup from "@views/Settings/TwoFactorAuthentication/TwoFactorAuthenticationOptionsMethodsRadioGroup";

vi.mock("react-i18next", () => ({
    useTranslation: () => ({ t: (key: string) => key }),
}));

vi.mock("@services/UserInfo", () => ({
    toMethod2FA: String,
}));

it("renders radio options for all provided methods", () => {
    render(
        <TwoFactorAuthenticationOptionsMethodsRadioGroup
            id="test"
            methods={[1, 2, 3]}
            method={1}
            name="Default Method"
            handleMethodChanged={vi.fn()}
        />,
    );
    expect(screen.getByText("WebAuthn")).toBeInTheDocument();
    expect(screen.getByText("One-Time Password")).toBeInTheDocument();
    expect(screen.getByText("Mobile Push")).toBeInTheDocument();
});

it("renders correct number of radio buttons", () => {
    render(
        <TwoFactorAuthenticationOptionsMethodsRadioGroup
            id="test"
            methods={[1, 2]}
            method={1}
            name="Default Method"
            handleMethodChanged={vi.fn()}
        />,
    );
    expect(screen.getAllByRole("radio")).toHaveLength(2);
});

it("reports the selected method as the value itself", () => {
    const handleMethodChanged = vi.fn();

    render(
        <TwoFactorAuthenticationOptionsMethodsRadioGroup
            id="test"
            methods={[1, 2]}
            method={1}
            name="Default Method"
            handleMethodChanged={handleMethodChanged}
        />,
    );

    fireEvent.click(screen.getAllByRole("radio")[1]);

    expect(handleMethodChanged).toHaveBeenCalledWith("2");
});
