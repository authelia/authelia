import React from "react";

import TextField, { TextFieldProps } from "@mui/material/TextField";

const OneTimeCodeTextField = function (props: TextFieldProps) {
    return (
        <TextField
            {...props}
            slotProps={{
                htmlInput: {
                    style: {
                        textTransform: "uppercase",
                        textAlign: "center",
                        letterSpacing: ".5rem",
                    },
                },
            }}
            variant={"outlined"}
            color={"info"}
            spellCheck={false}
        />
    );
};

export default OneTimeCodeTextField;
