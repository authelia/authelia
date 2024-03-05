import React from "react";

import { Theme } from "@mui/material";

declare module "@mui/material/styles" {
    interface Theme {
        custom: {
            icon: React.CSSProperties["color"];
            loadingBar: React.CSSProperties["color"];
        };
    }

    interface ThemeOptions {
        custom?: {
            icon?: React.CSSProperties["color"];
            loadingBar?: React.CSSProperties["color"];
        };
    }
}

declare module "@mui/styles/defaultTheme" {
    // eslint-disable-next-line @typescript-eslint/no-empty-interface
    interface DefaultTheme extends Theme {}
}

export const ThemeNameAuto = "auto";
export const ThemeNameLight = "light";
export const ThemeNameDark = "dark";
export const ThemeNameGrey = "grey";

export { default as Light } from "@themes/Light";
export { default as Dark } from "@themes/Dark";
export { default as Grey } from "@themes/Grey";
