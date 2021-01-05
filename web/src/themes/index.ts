declare module "@material-ui/core/styles/createMuiTheme" {
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

export { default as Light } from "./Light";
export { default as Dark } from "./Dark";
export { default as Grey } from "./Grey";
export { default as Custom } from "./Custom";
