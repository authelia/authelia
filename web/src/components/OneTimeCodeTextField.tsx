import React from "react";

import TextField, { TextFieldProps } from "@mui/material/TextField";

const OneTimeCodeTextField = function (props: TextFieldProps) {
    return (
        <TextField
            {...props}
            slotProps={{
                htmlInput: {
                    style: {
                        letterSpacing: ".5rem",
                        textAlign: "center",
                        textTransform: "uppercase",
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
