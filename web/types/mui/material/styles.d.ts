import { DeprecatedThemeOptions } from "@mui/material/styles";

declare module "@mui/material/styles" {
    export interface DeprecatedThemeOptions {
        custom: {
            icon: React.CSSProperties["color"];
            loadingBar: React.CSSProperties["color"];
        };
    }
    interface Theme {
        custom: {
            icon: React.CSSProperties["color"];
            loadingBar: React.CSSProperties["color"];
        };
    }
}
