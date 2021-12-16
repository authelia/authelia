import React from "react";

import { render } from "@testing-library/react";

import PasswordMeter from "@components/PasswordMeter";

it("renders without crashing", () => {
    render(<PasswordMeter value={""} classic minLength={4} />);
});

it("renders adjusted height without crashing", () => {
    render(<PasswordMeter value={"Passw0rd!"} minLength={4} />);
});
