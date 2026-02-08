import { createTheme } from "@mui/material/styles";

const Light = createTheme({
    custom: {
        icon: "#000",
        loadingBar: "#000",
    },
    palette: {
        action: {
            activatedOpacity: 0.12,
            active: "rgba(0, 0, 0, 0.54)",
            disabled: "rgba(0, 0, 0, 0.26)",
            disabledBackground: "rgba(0, 0, 0, 0.12)",
            disabledOpacity: 0.38,
            focus: "rgba(0, 0, 0, 0.12)",
            focusOpacity: 0.12,
            hover: "rgba(0, 0, 0, 0.04)",
            hoverOpacity: 0.04,
            selected: "rgba(0, 0, 0, 0.08)",
            selectedOpacity: 0.08,
        },
        background: {
            default: "#fff",
            paper: "#fff",
        },
        contrastThreshold: 3,
        divider: "rgba(0, 0, 0, 0.12)",
        error: {
            contrastText: "#ffffff",
            dark: "#d32f2f",
            light: "#e57373",
            main: "#f44336",
        },
        grey: {
            "100": "#f5f5f5",
            "200": "#eeeeee",
            "300": "#e0e0e0",
            "400": "#bdbdbd",
            "50": "#fafafa",
            "500": "#9e9e9e",
            "600": "#757575",
            "700": "#616161",
            "800": "#424242",
            "900": "#212121",
            A100: "#d5d5d5",
            A200: "#aaaaaa",
            A400: "#303030",
            A700: "#616161",
        },
        info: {
            contrastText: "#ffffff",
            dark: "#1976d2",
            light: "#64b5f6",
            main: "#2196f3",
        },
        mode: "light",
        primary: {
            main: "#1976d2",
        },
        secondary: {
            contrastText: "#ffffff",
            dark: "#c51162",
            light: "#ff4081",
            main: "#f50057",
        },
        success: {
            contrastText: "rgba(0, 0, 0, 0.87)",
            dark: "#388e3c",
            light: "#81c784",
            main: "#4caf50",
        },
        text: {
            disabled: "rgba(0, 0, 0, 0.38)",
            primary: "rgba(0, 0, 0, 0.87)",
            secondary: "rgba(0, 0, 0, 0.54)",
        },
        tonalOffset: 0.2,
        warning: {
            contrastText: "rgba(0, 0, 0, 0.87)",
            dark: "#f57c00",
            light: "#ffb74d",
            main: "#ff9800",
        },
    },
});

export default Light;
