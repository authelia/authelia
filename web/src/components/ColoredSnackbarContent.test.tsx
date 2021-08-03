import React from "react";

import { render, screen } from "@testing-library/react";
import ReactDOM from "react-dom";

import ColoredSnackbarContent from "@components/ColoredSnackbarContent";

it("renders without crashing", () => {
    const div = document.createElement("div");
    ReactDOM.render(<ColoredSnackbarContent level="success" message="this is a success" />, div);
    ReactDOM.unmountComponentAtNode(div);
});

it("should contain the message", () => {
    render(<ColoredSnackbarContent level="success" message="this is a success" />);
    expect(screen.getByRole("alert")).toHaveTextContent("this is a success");
});
