import { createMuiTheme } from "@material-ui/core/styles";

const Grey = createMuiTheme({
    custom: {
        icon: "#929aa5",
        loadingBar: "#929aa5",
    },
    palette: {
        primary: {
            main: "#929aa5",
        },
        background: {
            default: "#2f343e",
            paper: "#2f343e",
        },
    },
    overrides: {
        MuiCssBaseline: {
            "@global": {
                body: {
                    backgroundColor: "#2f343e",
                    color: "#929aa5",
                },
            },
        },
        MuiOutlinedInput: {
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
        MuiCheckbox: {
            root: {
                color: "#929aa5",
            },
        },
        MuiInputBase: {
            input: {
                color: "#929aa5",
            },
        },
        MuiInputLabel: {
            root: {
                color: "#929aa5",
            },
        },
    },
});

export default Grey;
