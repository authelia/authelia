import { createTheme } from "@mui/material/styles";

const Dark = createTheme({
    custom: {
        icon: "#fff",
        loadingBar: "#fff",
    },
    palette: {
        mode: "dark",
        primary: {
            main: "#1976d2",
        },
    },
});

export default Dark;
