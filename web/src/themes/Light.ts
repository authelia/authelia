import { createTheme } from "@material-ui/core/styles";

const Light = createTheme({
    custom: {
        icon: "#000",
        loadingBar: "#000",
    },
    palette: {
        primary: {
            main: "#1976d2",
        },
        background: {
            default: "#fff",
            paper: "#fff",
        },
    },
});

export default Light;
