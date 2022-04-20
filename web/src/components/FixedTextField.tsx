import React from "react";

import { Theme } from "@mui/material/styles";
import TextField, { TextFieldProps } from "@mui/material/TextField";
import makeStyles from "@mui/styles/makeStyles";

/**
 * This component fixes outlined TextField
 * https://github.com/mui-org/material-ui/issues/14530#issuecomment-463576879
 *
 * @param props the TextField props
 */
const FixedTextField = function (props: TextFieldProps) {
    const styles = useStyles();

    return (
        <TextField
            {...props}
            InputLabelProps={{
                classes: {
                    root: styles.label,
                },
            }}
        >
            {props.children}
        </TextField>
    );
};

export default FixedTextField;

const useStyles = makeStyles((theme: Theme) => ({
    label: {
        backgroundColor: theme.palette.background.default,
        paddingLeft: theme.spacing(0.1),
        paddingRight: theme.spacing(0.1),
    },
}));
