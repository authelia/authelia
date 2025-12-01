import { createTheme } from "@mui/material/styles";

const Oled = createTheme({
    custom: {
        icon: "#fff",
        loadingBar: "#2aa2c1",
    },
    palette: {
        action: {
            activatedOpacity: 0.24,
            active: "#2aa2c1",
            disabled: "rgba(255, 255, 255, 0.3)",
            disabledBackground: "rgba(255, 255, 255, 0.12)",
            disabledOpacity: 0.5,
            focus: "rgba(42, 162, 193, 0.12)",
            focusOpacity: 0.12,
            hover: "hsla(206, 100%, 50%, 0.04)",
            hoverOpacity: 0.08,
            selected: "rgba(42, 162, 193, 0.16)",
            selectedOpacity: 0.16,
        },
        background: {
            default: "#000000",
            paper: "#0a0a0a",
        },
        contrastThreshold: 3,
        divider: "#1a1a1a",
        error: {
            contrastText: "#ffffff",
            dark: "#991b1b",
            light: "#ff6b6b",
            main: "#dc3545",
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
            main: "#2aa2c1",
            dark: "#238a9f",
            light: "#64b5f6",
            contrastText: "#ffffff",
        },
        secondary: {
            contrastText: "#ffffff",
            dark: "#1a1a1a",
            light: "#2a2a2a",
            main: "#1a1a1a",
        },
        success: {
            contrastText: "#ffffff",
            dark: "#30a46c",
            light: "#81c784",
            main: "#30a46c",
        },
        text: {
            disabled: "rgba(255, 255, 255, 0.5)",
            primary: "hsla(0, 0%, 100%, 0.74)",
            secondary: "hsla(0, 0%, 100%, 0.51)",
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
