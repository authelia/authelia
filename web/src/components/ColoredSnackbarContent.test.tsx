import React from "react";

import { createTheme } from "@mui/material/styles";
import { ThemeProvider } from "@mui/styles";
import { render, screen } from "@testing-library/react";

import ColoredSnackbarContent from "@components/ColoredSnackbarContent";

it("renders without crashing", () => {
    render(
        <ThemeProvider theme={createTheme()}>
            <ColoredSnackbarContent level="success" message="this is a success" />
            );
        </ThemeProvider>,
    );
});

it("should contain the message", () => {
    render(
        <ThemeProvider theme={createTheme()}>
            <ColoredSnackbarContent level="success" message="this is a success" />
        </ThemeProvider>,
    );
    expect(screen.getByRole("alert")).toHaveTextContent("this is a success");
});
