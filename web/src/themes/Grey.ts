import { createTheme, adaptV4Theme } from "@mui/material/styles";

const Grey = createTheme({
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
});

export default Grey;
