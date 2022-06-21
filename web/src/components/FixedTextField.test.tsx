import React from "react";

import { createTheme } from "@mui/material/styles";
import { ThemeProvider } from "@mui/styles";
import { render } from "@testing-library/react";

import FixedTextField from "@components/FixedTextField";

it("renders without crashing", () => {
    render(
        <ThemeProvider theme={createTheme()}>
            <FixedTextField />
        </ThemeProvider>,
    );
});
