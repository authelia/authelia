import { createTheme } from "@mui/material/styles";

const Oled = createTheme({
    custom: {
        icon: "#fff",
        loadingBar: "#fff",
    },
    palette: {
        action: {
            activatedOpacity: 0.24,
            active: "#ffffff",
            disabled: "rgba(255, 255, 255, 0.3)",
            disabledBackground: "rgba(255, 255, 255, 0.12)",
            disabledOpacity: 0.38,
            focus: "rgba(255, 255, 255, 0.12)",
            focusOpacity: 0.12,
            hover: "rgba(255, 255, 255, 0.08)",
            hoverOpacity: 0.08,
            selected: "rgba(255, 255, 255, 0.16)",
            selectedOpacity: 0.16,
        },
        background: {
            default: "#000000",
            paper: "#000000",
        },
        contrastThreshold: 3,
        divider: "rgba(255, 255, 255, 0.12)",
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
        mode: "dark",
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
            contrastText: "rgba(255, 255, 255, 0.87)",
            dark: "#388e3c",
            light: "#81c784",
            main: "#4caf50",
        },
        text: {
            disabled: "rgba(255, 255, 255, 0.5)",
            primary: "#ffffff",
            secondary: "rgba(255, 255, 255, 0.7)",
        },
        tonalOffset: 0.2,
        warning: {
            contrastText: "rgba(255, 255, 255, 0.87)",
            dark: "#f57c00",
            light: "#ffb74d",
            main: "#ff9800",
        },
    },
});

export default Oled;
