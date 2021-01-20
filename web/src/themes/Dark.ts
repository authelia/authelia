import { createMuiTheme } from "@material-ui/core/styles";

const Dark = createMuiTheme({
    custom: {
        icon: "#fff",
        loadingBar: "#fff",
    },
    palette: {
        type: "dark",
        primary: {
            main: "#1976d2",
        },
    },
});

export default Dark;
