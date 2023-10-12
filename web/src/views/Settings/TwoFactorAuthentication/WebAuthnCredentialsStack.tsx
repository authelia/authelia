import React from "react";

import Grid from "@mui/material/Unstable_Grid2";

import { WebAuthnCredential } from "@models/WebAuthn";
import WebAuthnCredentialItem from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialItem";

interface Props {
    credentials: WebAuthnCredential[];
    handleRefreshState: () => void;
    handleDelete: (index: number) => void;
    handleEdit: (index: number) => void;
}

const WebAuthnCredentialsStack = function (props: Props) {
    return (
        <Grid container spacing={3}>
            {props.credentials.map((x, idx) => (
                <WebAuthnCredentialItem
                    key={idx}
                    index={idx}
                    credential={x}
                    handleDelete={props.handleDelete}
                    handleEdit={props.handleEdit}
                />
            ))}
        </Grid>
    );
};

export default WebAuthnCredentialsStack;
