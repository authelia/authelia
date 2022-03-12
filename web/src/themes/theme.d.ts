import React from "react";

declare module "@mui/material/styles" {
    interface ThemeOptions {
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

declare module "@mui/styles/defaultTheme" {
    interface DefaultTheme extends Theme {}
}
