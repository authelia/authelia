import { render, screen } from "@testing-library/react";

import PasswordMeter from "@components/PasswordMeter";
import "@i18n/index";
import { PasswordPolicyMode } from "@models/PasswordPolicy";

it("renders without crashing", () => {
    render(
        <PasswordMeter
            value={""}
            policy={{
                max_length: 0,
                min_length: 4,
                min_score: 0,
                mode: PasswordPolicyMode.Standard,
                require_lowercase: false,
                require_number: false,
                require_special: false,
                require_uppercase: false,
            }}
        />,
    );
});

it("renders adjusted height without crashing", () => {
    render(
        <PasswordMeter
            value={"Passw0rd!"}
            policy={{
                max_length: 0,
                min_length: 4,
                min_score: 0,
                mode: PasswordPolicyMode.Standard,
                require_lowercase: false,
                require_number: false,
                require_special: false,
                require_uppercase: false,
            }}
        />,
    );
});

it("displays warning message on password too large", async () => {
    const maxLength = 5;
    render(
        <PasswordMeter
            value={"password"}
            policy={{
                max_length: maxLength,
                min_length: 4,
                min_score: 0,
                mode: PasswordPolicyMode.Standard,
                require_lowercase: true,
                require_number: true,
                require_special: true,
                require_uppercase: true,
            }}
        />,
    );

    const text = `Must not be more than ${maxLength} characters in length`;
    expect(screen.queryByText(text)).toBeInTheDocument();
});

it("displays warning message on password too short", async () => {
    const minLength = 5;
    render(
        <PasswordMeter
            value={"abc"}
            policy={{
                max_length: 0,
                min_length: minLength,
                min_score: 0,
                mode: PasswordPolicyMode.Standard,
                require_lowercase: true,
                require_number: true,
                require_special: true,
                require_uppercase: true,
            }}
        />,
    );

    const text = `Must be at least ${minLength} characters in length`;
    expect(screen.queryByText(text)).toBeInTheDocument();
});

it("displays warning message on password policy fail", async () => {
    render(
        <PasswordMeter
            value={""}
            policy={{
                max_length: 0,
                min_length: 0,
                min_score: 0,
                mode: PasswordPolicyMode.Standard,
                require_lowercase: true,
                require_number: true,
                require_special: true,
                require_uppercase: true,
            }}
        />,
    );

    expect(screen.queryByText("The password does not meet the password policy")).toBeInTheDocument();
});
