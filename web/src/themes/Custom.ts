import { createMuiTheme } from "@material-ui/core/styles";

import { getPrimaryColor, getSecondaryColor } from "../utils/Configuration";

const Custom = createMuiTheme({
    custom: {
        icon: getPrimaryColor(),
        loadingBar: getPrimaryColor(),
    },
    palette: {
        primary: {
            main: getPrimaryColor(),
        },
        background: {
            default: getSecondaryColor(),
            paper: getSecondaryColor(),
        },
    },
    overrides: {
        MuiCssBaseline: {
            "@global": {
                body: {
                    backgroundColor: getSecondaryColor(),
                    color: getPrimaryColor(),
                },
            },
        },
        MuiOutlinedInput: {
            root: {
                "& $notchedOutline": {
                    borderColor: getPrimaryColor(),
                },
                "&:hover:not($disabled):not($focused):not($error) $notchedOutline": {
                    borderColor: getPrimaryColor(),
                    borderWidth: 2,
                },
                "&$focused $notchedOutline": {
                    borderColor: getPrimaryColor(),
                },
            },
            notchedOutline: {},
        },
        MuiCheckbox: {
            root: {
                color: getPrimaryColor(),
            },
        },
        MuiInputBase: {
            input: {
                color: getPrimaryColor(),
            },
        },
        MuiInputLabel: {
            root: {
                color: getPrimaryColor(),
            },
        },
        MuiButton: {
            label: {
                color: getSecondaryColor(),
            },
        },
    },
});

export default Custom;
