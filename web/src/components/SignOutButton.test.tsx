import React from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import SignOutButton from "@components/SignOutButton";

const mockDoSignOut = vi.fn();

vi.mock("@hooks/SignOut", () => ({
    useSignOut: vi.fn(() => mockDoSignOut),
}));

it("renders without crashing", () => {
    render(<SignOutButton id="test" text="Sign Out" />);
});

it("renders button with translated text", () => {
    render(<SignOutButton id="test" text="Sign Out" />);
    expect(screen.getByText("Sign Out")).toBeInTheDocument();
});

it("renders tooltip when provided", () => {
    render(<SignOutButton id="test" text="Sign Out" tooltip="Sign out" />);
    expect(screen.getByText("Sign Out")).toBeInTheDocument();
});

it("calls sign out on click", () => {
    render(<SignOutButton id="test" text="Sign Out" />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockDoSignOut).toHaveBeenCalledWith(false);
});

it("calls sign out with preserve", () => {
    render(<SignOutButton id="test" text="Sign Out" preserve={true} />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockDoSignOut).toHaveBeenCalledWith(true);
});
