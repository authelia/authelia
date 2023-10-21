import React from "react";

import Grid from "@mui/material/Unstable_Grid2/Grid2";

import { WebAuthnCredential } from "@models/WebAuthn";
import WebAuthnCredentialItem from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialItem";

interface Props {
    credentials: WebAuthnCredential[];
    handleRefreshState: () => void;
    handleInformation: (index: number) => void;
    handleEdit: (index: number) => void;
    handleDelete: (index: number) => void;
}

const WebAuthnCredentialsGrid = function (props: Props) {
    return (
        <Grid container spacing={3}>
            {props.credentials.map((credential, index) => (
                <Grid xs={12} md={6} xl={3} key={index}>
                    <WebAuthnCredentialItem
                        index={index}
                        credential={credential}
                        handleInformation={props.handleInformation}
                        handleEdit={props.handleEdit}
                        handleDelete={props.handleDelete}
                    />
                </Grid>
            ))}
        </Grid>
    );
};

export default WebAuthnCredentialsGrid;
