import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import PasswordMeter from "@components/PasswordMeter";
import "@i18n/index";
import { PasswordPolicyMode } from "@models/PasswordPolicy";

vi.mock("zxcvbn", () => ({ default: vi.fn(() => ({ score: 3, feedback: { warning: "Test warning" } })) }));

it("renders without crashing", () => {
    render(
        <PasswordMeter
            value={""}
            policy={{
                max_length: 0,
                min_length: 4,
                min_score: 0,
                require_lowercase: false,
                require_number: false,
                require_special: false,
                require_uppercase: false,
                mode: PasswordPolicyMode.Standard,
            }}
        />,
    );
});

it("displays warning message on password too long", () => {
    const maxLength = 5;
    render(
        <PasswordMeter
            value={"password"}
            policy={{
                max_length: maxLength,
                min_length: 4,
                min_score: 0,
                require_lowercase: true,
                require_number: true,
                require_special: true,
                require_uppercase: true,
                mode: PasswordPolicyMode.Standard,
            }}
        />,
    );

    const text = `Must not be more than ${maxLength} characters in length`;
    expect(screen.getByText(text)).toBeInTheDocument();
});

it("displays warning message on password too short", () => {
    const minLength = 5;
    render(
        <PasswordMeter
            value={"abc"}
            policy={{
                max_length: 0,
                min_length: minLength,
                min_score: 0,
                require_lowercase: true,
                require_number: true,
                require_special: true,
                require_uppercase: true,
                mode: PasswordPolicyMode.Standard,
            }}
        />,
    );

    const text = `Must be at least ${minLength} characters in length`;
    expect(screen.getByText(text)).toBeInTheDocument();
});

it("displays warning message on password policy fail for missing lowercase", () => {
    render(
        <PasswordMeter
            value={"PASSWORD123!"}
            policy={{
                max_length: 0,
                min_length: 0,
                min_score: 0,
                require_lowercase: true,
                require_number: false,
                require_special: false,
                require_uppercase: false,
                mode: PasswordPolicyMode.Standard,
            }}
        />,
    );

    expect(screen.getByText("The password does not meet the password policy")).toBeInTheDocument();
    expect(
        screen.getByText((content) => content.includes("Must have at least one lowercase letter")),
    ).toBeInTheDocument();
});

it("displays warning message on password policy fail for missing uppercase, number and special character", () => {
    render(
        <PasswordMeter
            value={"password"}
            policy={{
                max_length: 0,
                min_length: 0,
                min_score: 0,
                require_lowercase: false,
                require_number: true,
                require_special: true,
                require_uppercase: true,
                mode: PasswordPolicyMode.Standard,
            }}
        />,
    );

    expect(screen.getByText("The password does not meet the password policy")).toBeInTheDocument();
    expect(
        screen.getByText((content) => content.includes("* Must have at least one UPPERCASE letter")),
    ).toBeInTheDocument();
    expect(screen.getByText((content) => content.includes("* Must have at least one number"))).toBeInTheDocument();
    expect(
        screen.getByText((content) => content.includes("* Must have at least one special character")),
    ).toBeInTheDocument();
});

it("uses ZXCVBN mode", () => {
    render(
        <PasswordMeter
            value={"password"}
            policy={{
                max_length: 0,
                min_length: 0,
                min_score: 0,
                require_lowercase: false,
                require_number: false,
                require_special: false,
                require_uppercase: false,
                mode: PasswordPolicyMode.ZXCVBN,
            }}
        />,
    );

    expect(screen.getByText("Test warning")).toBeInTheDocument();
});

it("does not display warning when password meets policy", () => {
    render(
        <PasswordMeter
            value={"Password1!"}
            policy={{
                max_length: 0,
                min_length: 8,
                min_score: 0,
                require_lowercase: true,
                require_number: true,
                require_special: true,
                require_uppercase: true,
                mode: PasswordPolicyMode.Standard,
            }}
        />,
    );

    expect(screen.queryByText("The password does not meet the password policy")).toBeNull();
});
