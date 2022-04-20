import React from "react";

import { Theme, useTheme } from "@mui/material";
import TextField, { TextFieldProps } from "@mui/material/TextField";
import { CSSProperties } from "@mui/styles";

/**
 * This component fixes outlined TextField
 * https://github.com/mui-org/material-ui/issues/14530#issuecomment-463576879
 *
 * @param props the TextField props
 */
const FixedTextField = function (props: TextFieldProps) {
    const theme = useTheme();
    const styles = useStyles(theme);

    return (
        <TextField
            {...props}
            InputLabelProps={{
                sx: styles.label,
            }}
        >
            {props.children}
        </TextField>
    );
};

export default FixedTextField;

const useStyles = (theme: Theme): { [key: string]: CSSProperties } => ({
    label: {
        backgroundColor: theme.palette.background.default,
        paddingLeft: theme.spacing(0.1),
        paddingRight: theme.spacing(0.1),
    },
});
