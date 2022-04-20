import { createTheme } from "@mui/material/styles";

const Light = createTheme({
    custom: {
        icon: "#000",
        loadingBar: "#000",
    },
    palette: {
        mode: "light",
        primary: {
            light: "#a6d4fa",
            main: "#1976d2",
            dark: "#648dae",
        },
        secondary: {
            light: "#f6a5c0",
            main: "#f48fb1",
            dark: "#aa647b",
        },
        error: {
            light: "#e57373",
            main: "#f44336",
            dark: "#d32f2f",
        },
        warning: {
            light: "#ffb74d",
            main: "#ff9800",
            dark: "#f57c00",
        },
        info: {
            light: "#81c784",
            main: "#4caf50",
            dark: "#388e3c",
        },
        text: {
            primary: "rgba(0, 0, 0, 0.87)",
            secondary: "rgba(0, 0, 0, 0.54)",
            disabled: "rgba(0, 0, 0, 0.38)",
        },
        action: {
            active: "rgba(0, 0, 0, 0.54)",
            hover: "rgba(0, 0, 0, 0.04)",
            selected: "rgba(0, 0, 0, 0.08)",
            disabled: "rgba(0, 0, 0, 0.26)",
            disabledBackground: "rgba(0, 0, 0, 0.12)",
        },
        background: {
            default: "#fafafa",
            paper: "#fff",
        },
        divider: "rgba(0, 0, 0, 0.12)",
    },
});

export default Light;
