import { createTheme, adaptV4Theme } from "@mui/material/styles";

const Dark = createTheme(
    adaptV4Theme({
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
    }),
);

export default Dark;
