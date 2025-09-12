import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import SignOutButton from "@components/SignOutButton";
import SwitchUserButton from "@components/SwitchUserButton";

vi.mock("@components/SignOutButton", () => ({
    default: vi.fn(() => <button>Switch User</button>),
}));

it("renders without crashing", () => {
    render(<SwitchUserButton />);
});

it("renders switch user button", () => {
    render(<SwitchUserButton />);
    expect(screen.getByText("Switch User")).toBeInTheDocument();
});

it("passes correct props to sign out button", () => {
    render(<SwitchUserButton />);
    expect(SignOutButton).toHaveBeenCalledWith(
        {
            id: "switch-user-button",
            text: "Switch User",
            tooltip: "Logout and continue the current flow",
            preserve: true,
        },
        undefined,
    );
});
