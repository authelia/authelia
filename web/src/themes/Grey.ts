import { createTheme } from "@mui/material/styles";

const Grey = createTheme({
    components: {
        MuiCheckbox: {
            styleOverrides: {
                root: {
                    color: "#929aa5",
                },
            },
        },
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
        MuiOutlinedInput: {
            styleOverrides: {
                notchedOutline: {},
                root: {
                    "&:hover:not($disabled):not($focused):not($error) $notchedOutline": {
                        borderColor: "#929aa5",
                        borderWidth: 2,
                    },
                    "&$focused $notchedOutline": {
                        borderColor: "#929aa5",
                    },
                    "& $notchedOutline": {
                        borderColor: "#929aa5",
                    },
                },
            },
        },
    },
    custom: {
        icon: "#929aa5",
        loadingBar: "#929aa5",
    },
    palette: {
        action: {
            activatedOpacity: 0.12,
            active: "#929aa5",
            disabled: "rgba(0, 0, 0, 0.26)",
            disabledBackground: "rgba(0, 0, 0, 0.12)",
            disabledOpacity: 0.38,
            focus: "rgba(0, 0, 0, 0.12)",
            focusOpacity: 0.12,
            hover: "#929aa5",
            hoverOpacity: 0.04,
            selected: "#929aa5",
            selectedOpacity: 0.08,
        },
        background: {
            default: "#2f343e",
            paper: "#2f343e",
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
        mode: "dark",
        primary: {
            main: "#929aa5",
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
            primary: "#929aa5",
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

export default Grey;
