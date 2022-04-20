import { createTheme } from "@mui/material/styles";

const Dark = createTheme({
    custom: {
        icon: "#fff",
        loadingBar: "#fff",
    },
    palette: {
        mode: "dark",
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
            primary: "#fff",
            secondary: "rgba(255, 255, 255, 0.7)",
            disabled: "rgba(255, 255, 255, 0.5)",
        },
        action: {
            active: "#fff",
            hover: "rgba(255, 255, 255, 0.08)",
            selected: "rgba(255, 255, 255, 0.16)",
            disabled: "rgba(255, 255, 255, 0.3)",
            disabledBackground: "rgba(255, 255, 255, 0.12)",
        },
        background: {
            default: "#303030",
            paper: "#424242",
        },
        divider: "rgba(255, 255, 255, 0.12)",
    },
});

export default Dark;
