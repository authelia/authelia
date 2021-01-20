import React from "react";

import { makeStyles } from "@material-ui/core";
import TextField, { TextFieldProps } from "@material-ui/core/TextField";

/**
 * This component fixes outlined TextField
 * https://github.com/mui-org/material-ui/issues/14530#issuecomment-463576879
 *
 * @param props the TextField props
 */
const FixedTextField = function (props: TextFieldProps) {
    const style = useStyles();
    return (
        <TextField
            {...props}
            InputLabelProps={{
                classes: {
                    root: style.label,
                },
            }}
            inputProps={{ autoCapitalize: props.autoCapitalize }}
        >
            {props.children}
        </TextField>
    );
};

export default FixedTextField;

const useStyles = makeStyles((theme) => ({
    label: {
        backgroundColor: theme.palette.background.default,
        paddingLeft: theme.spacing(0.1),
        paddingRight: theme.spacing(0.1),
    },
}));
