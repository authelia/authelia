import React from "react";

import { useTheme } from "@mui/material";
import { Theme } from "@mui/material/styles";
import TextField, { TextFieldProps } from "@mui/material/TextField";

import { StylesProperties } from "@models/StylesProperties";

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

const useStyles = (theme: Theme): StylesProperties => ({
    label: {
        paddingLeft: theme.spacing(0.1),
        paddingRight: theme.spacing(0.1),
    },
});
