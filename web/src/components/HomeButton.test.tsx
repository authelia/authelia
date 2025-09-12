import React from "react";

import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";

import HomeButton from "@components/HomeButton";

const mockNavigate = vi.fn();

vi.mock("@hooks/RouterNavigate", () => ({
    useRouterNavigate: vi.fn(() => mockNavigate),
}));

vi.mock("@constants/Routes", () => ({
    IndexRoute: "/",
}));

it("renders without crashing", () => {
    render(<HomeButton />);
});

it("renders button with home text", () => {
    render(<HomeButton />);
    expect(screen.getByText("Home")).toBeInTheDocument();
});

it("navigates on click", () => {
    render(<HomeButton />);
    const button = screen.getByRole("button");
    fireEvent.click(button);
    expect(mockNavigate).toHaveBeenCalledWith("/", false, false, false);
});
