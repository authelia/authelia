import React from "react";

import TextField, { TextFieldProps } from "@mui/material/TextField";

const OneTimeCodeTextField = function (props: TextFieldProps) {
    return (
        <TextField
            {...props}
            inputProps={{
                style: {
                    textTransform: "uppercase",
                    textAlign: "center",
                    letterSpacing: ".5rem",
                },
            }}
            variant={"outlined"}
            spellCheck={false}
        />
    );
};

export default OneTimeCodeTextField;
