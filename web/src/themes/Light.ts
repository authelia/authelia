import { createTheme } from "@mui/material/styles";

const Light = createTheme({
    custom: {
        icon: "#2aa2c1",
        loadingBar: "#2aa2c1",
    },
    typography: {
        fontFamily: "system-ui, -apple-system, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif",
        h1: {
            fontWeight: 700,
            fontSize: '2.25rem',
            color: '#FFFFFF',
        },
        h2: {
            fontWeight: 700,
            fontSize: '1.5rem',
            color: '#FFFFFF',
        },
        h5: {
            fontWeight: 600,
            fontSize: '1.25rem',
            color: '#FFFFFF',
        },
        h6: {
            fontWeight: 600,
            fontSize: '1rem',
            color: 'hsla(0, 0%, 100%, 0.74)',
        },
        body1: {
            fontSize: '0.875rem',
            color: 'hsla(0, 0%, 100%, 0.74)',
        },
        button: {
            textTransform: 'none',
            fontWeight: 600,
        },
    },
    palette: {
        action: {
            activatedOpacity: 0.12,
            active: "rgba(42, 162, 193, 0.54)",
            disabled: "rgba(255, 255, 255, 0.26)",
            disabledBackground: "rgba(255, 255, 255, 0.12)",
            disabledOpacity: 0.38,
            focus: "rgba(42, 162, 193, 0.12)",
            focusOpacity: 0.12,
            hover: "hsla(206, 100%, 50%, 0.04)",
            hoverOpacity: 0.04,
            selected: "rgba(42, 162, 193, 0.08)",
            selectedOpacity: 0.08,
        },
        background: {
            default: "#081727",
            paper: "#1e2b39",
        },
        contrastThreshold: 3,
        divider: "#2f3d4d",
        primary: {
            main: "#2aa2c1",
            light: "#3ab5d4",
            dark: "#238a9f",
            contrastText: "#ffffff",
        },
        secondary: {
            main: "#2aa2c1",
            light: "#3ab5d4",
            dark: "#238a9f",
            contrastText: "#ffffff",
        },
        error: {
            contrastText: "#ffffff",
            dark: "#991b1b",
            light: "#ff6b6b",
            main: "#dc3545",
        },
        success: {
            contrastText: "#ffffff",
            dark: "#2d8659",
            light: "#4aba7f",
            main: "#30a46c",
        },
        text: {
            primary: "hsla(0, 0%, 100%, 0.74)",
            secondary: "hsla(0, 0%, 100%, 0.51)",
            disabled: "hsla(0, 0%, 100%, 0.38)",
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
        warning: {
            contrastText: "#ffffff",
            dark: "#f57c00",
            light: "#ffb74d",
            main: "#ff9800",
        },
        mode: "dark",
        tonalOffset: 0.2,
    },
});

export default Light;
