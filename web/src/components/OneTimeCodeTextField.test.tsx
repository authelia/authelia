import { ThemeProvider, createTheme } from "@mui/material/styles";
import { render } from "@testing-library/react";

import OneTimeCodeTextField from "@components/OneTimeCodeTextField";

it("renders without crashing", () => {
    render(
        <ThemeProvider theme={createTheme()}>
            <OneTimeCodeTextField />
        </ThemeProvider>,
    );
});
