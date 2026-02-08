import { CSSProperties } from "react";

/* eslint-disable no-unused-vars */
declare module "@mui/material/styles" {
    interface Theme {
        custom: {
            icon: CSSProperties["color"];
            loadingBar: CSSProperties["color"];
        };
    }

    interface ThemeOptions {
        custom?: {
            icon?: CSSProperties["color"];
            loadingBar?: CSSProperties["color"];
        };
    }
}

export const ThemeNameAuto = "auto";
export const ThemeNameLight = "light";
export const ThemeNameDark = "dark";
export const ThemeNameGrey = "grey";
export const ThemeNameOled = "oled";

export { default as Light } from "@themes/Light";
export { default as Dark } from "@themes/Dark";
export { default as Grey } from "@themes/Grey";
export { default as Oled } from "@themes/Oled";
