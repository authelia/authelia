declare module "@material-ui/core/styles/createTheme" {
    interface Theme {
        custom: {
            icon: React.CSSProperties["color"];
            loadingBar: React.CSSProperties["color"];
        };
    }
    interface ThemeOptions {
        custom: {
            icon: React.CSSProperties["color"];
            loadingBar: React.CSSProperties["color"];
        };
    }
}

export { default as Light } from "@themes/Light";
export { default as Dark } from "@themes/Dark";
export { default as Grey } from "@themes/Grey";
