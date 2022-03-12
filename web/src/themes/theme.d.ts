import React from "react";

import { ThemeOptions, Theme } from "@mui/material/styles";

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
