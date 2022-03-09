import { createTheme, adaptV4Theme } from "@mui/material/styles";

const Light = createTheme(
    adaptV4Theme({
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
    }),
);

export default Light;
