import { createTheme } from "@mui/material/styles";

const Grey = createTheme({
    custom: {
        icon: "#929aa5",
        loadingBar: "#929aa5",
    },
    palette: {
        mode: "dark",
        primary: {
            main: "#929aa5",
        },
        secondary: {
            light: "#ff4081",
            main: "#f50057",
            dark: "#c51162",
            contrastText: "#ffffff",
        },
        error: {
            light: "#e57373",
            main: "#f44336",
            dark: "#d32f2f",
            contrastText: "#ffffff",
        },
        warning: {
            light: "#ffb74d",
            main: "#ff9800",
            dark: "#f57c00",
            contrastText: "rgba(0, 0, 0, 0.87)",
        },
        info: {
            light: "#64b5f6",
            main: "#2196f3",
            dark: "#1976d2",
            contrastText: "#ffffff",
        },
        success: {
            light: "#81c784",
            main: "#4caf50",
            dark: "#388e3c",
            contrastText: "rgba(0, 0, 0, 0.87)",
        },
        grey: {
            "50": "#fafafa",
            "100": "#f5f5f5",
            "200": "#eeeeee",
            "300": "#e0e0e0",
            "400": "#bdbdbd",
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
        contrastThreshold: 3,
        tonalOffset: 0.2,
        text: {
            primary: "#929aa5",
            secondary: "rgba(0, 0, 0, 0.54)",
            disabled: "rgba(0, 0, 0, 0.38)",
        },
        divider: "rgba(0, 0, 0, 0.12)",
        background: {
            paper: "#2f343e",
            default: "#2f343e",
        },
        action: {
            active: "#929aa5",
            hover: "#929aa5",
            hoverOpacity: 0.04,
            selected: "#929aa5",
            selectedOpacity: 0.08,
            disabled: "rgba(0, 0, 0, 0.26)",
            disabledBackground: "rgba(0, 0, 0, 0.12)",
            disabledOpacity: 0.38,
            focus: "rgba(0, 0, 0, 0.12)",
            focusOpacity: 0.12,
            activatedOpacity: 0.12,
        },
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
