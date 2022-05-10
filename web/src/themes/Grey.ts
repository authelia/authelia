import { createTheme } from "@mui/material/styles";

const Grey = createTheme({
    custom: {
        icon: "#929aa5",
        loadingBar: "#929aa5",
    },
    palette: {
        mode: "dark",
        primary: {
            light: "#a6d4fa",
            main: "#929aa5",
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
            primary: "#929aa5",
            secondary: "rgba(0, 0, 0, 0.54)",
            disabled: "rgba(0, 0, 0, 0.38)",
        },
        action: {
            active: "#929aa5",
            hover: "#929aa5",
            selected: "#929aa5",
            disabled: "rgba(0, 0, 0, 0.26)",
            disabledBackground: "rgba(0, 0, 0, 0.12)",
        },
        background: {
            default: "#2f343e",
            paper: "#2f343e",
        },
        divider: "rgba(0, 0, 0, 0.12)",
    },
    components: {
        MuiCssBaseline: {
            styleOverrides: {
                "@global": {
                    body: {
                        backgroundColor: "#2f343e",
                        color: "#929aa5",
                    },
                },
            },
        },
        MuiOutlinedInput: {
            styleOverrides: {
                root: {
                    "& $notchedOutline": {
                        borderColor: "#929aa5",
                    },
                    "&:hover:not($disabled):not($focused):not($error) $notchedOutline": {
                        borderColor: "#929aa5",
                        borderWidth: 2,
                    },
                    "&$focused $notchedOutline": {
                        borderColor: "#929aa5",
                    },
                },
                notchedOutline: {},
            },
        },
        MuiCheckbox: {
            styleOverrides: {
                root: {
                    color: "#929aa5",
                },
            },
        },
        MuiInputBase: {
            styleOverrides: {
                input: {
                    color: "#929aa5",
                },
            },
        },
        MuiInputLabel: {
            styleOverrides: {
                root: {
                    color: "#929aa5",
                },
            },
        },
    },
});

export default Grey;
