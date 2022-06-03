import React from "react";

import { render, screen } from "@testing-library/react";

import ColoredSnackbarContent from "@components/ColoredSnackbarContent";

it("renders without crashing", () => {
    render(<ColoredSnackbarContent level="success" message="" />);
    expect(screen.getByRole("alert")).toHaveTextContent("");
});

it("should contain the message", () => {
    render(<ColoredSnackbarContent level="success" message="this is a success" />);
    expect(screen.getByRole("alert")).toHaveTextContent("this is a success");
});
