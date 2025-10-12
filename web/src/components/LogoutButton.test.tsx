import React from "react";

import { render, screen } from "@testing-library/react";
import { vi } from "vitest";

import LogoutButton from "@components/LogoutButton";
import SignOutButton from "@components/SignOutButton";

vi.mock("@components/SignOutButton", () => ({
    default: vi.fn(() => <button>Logout</button>),
}));

it("renders without crashing", () => {
    render(<LogoutButton />);
});

it("renders button with correct props", () => {
    render(<LogoutButton />);
    expect(SignOutButton).toHaveBeenCalledWith(
        {
            id: "logout-button",
            text: "Logout",
            tooltip: "Logout and clear any current flow",
        },
        undefined,
    );
});

it("renders logout button", () => {
    render(<LogoutButton />);
    expect(screen.getByText("Logout")).toBeInTheDocument();
});
