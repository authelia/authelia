import React, { CSSProperties, ReactNode } from "react";

import { Box, Button, Grid, Theme, Typography, useTheme } from "@mui/material";

interface Props {
    id: string;
    method: string;
    icon: ReactNode;

    onClick: () => void;
}

function MethodItem(props: Props) {
    const theme = useTheme();
    const style = useStyles(theme);

    return (
        <Grid item xs={12} className="method-option" id={props.id}>
            <Button
                sx={{ ...style.item, ...style.buttonRoot }}
                color="primary"
                variant="contained"
                onClick={props.onClick}
            >
                <Box sx={style.icon}>{props.icon}</Box>
                <Box>
                    <Typography>{props.method}</Typography>
                </Box>
            </Button>
        </Grid>
    );
}

export default MethodItem;

const useStyles = (theme: Theme): { [key: string]: CSSProperties } => ({
    item: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
        width: "100%",
    },
    icon: {
        display: "inline-block",
        fill: "white",
    },
    buttonRoot: {
        display: "block",
    },
});
