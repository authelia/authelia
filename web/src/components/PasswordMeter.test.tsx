import React from "react";

import { render } from "@testing-library/react";

import PasswordMeter from "@components/PasswordMeter";
import "@i18n/index.ts";
import { PasswordPolicyMode } from "@models/PasswordPolicy";

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

it("renders adjusted height without crashing", () => {
    render(
        <PasswordMeter
            value={"Passw0rd!"}
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
