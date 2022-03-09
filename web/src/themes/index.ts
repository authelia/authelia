declare module "@mui/material/styles/createTheme" {
    interface Theme {
        custom: {
            icon: React.CSSProperties["color"];
            loadingBar: React.CSSProperties["color"];
        };
    }
    interface DeprecatedThemeOptions {
        custom: {
            icon: React.CSSProperties["color"];
            loadingBar: React.CSSProperties["color"];
        };
    }
}

export { default as Light } from "@themes/Light";
export { default as Dark } from "@themes/Dark";
export { default as Grey } from "@themes/Grey";
