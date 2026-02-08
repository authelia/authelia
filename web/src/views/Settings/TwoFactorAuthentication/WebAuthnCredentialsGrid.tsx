import Grid from "@mui/material/Grid";

import { WebAuthnCredential } from "@models/WebAuthn";
import WebAuthnCredentialItem from "@views/Settings/TwoFactorAuthentication/WebAuthnCredentialItem";

interface Props {
    credentials: WebAuthnCredential[];
    handleInformation: (_index: number) => void;
    handleEdit: (_index: number) => void;
    handleDelete: (_index: number) => void;
}

const WebAuthnCredentialsGrid = function (props: Props) {
    return (
        <Grid container spacing={3}>
            {props.credentials.map((credential, index) => (
                <Grid size={{ md: 6, xl: 3, xs: 12 }} key={credential.id}>
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
