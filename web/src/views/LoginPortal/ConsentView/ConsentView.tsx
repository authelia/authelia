import React, { useEffect } from "react";
import { Grid, Button } from "@material-ui/core";
import LoginLayout from "../../../layouts/LoginLayout";
import { useRequestedScopes } from "../../../hooks/Consent";
import { useNotifications } from "../../../hooks/NotificationsContext";

export interface Props { }

export default function (props: Props) {
    const { createErrorNotification } = useNotifications();
    const [resp, fetch, , err] = useRequestedScopes();

    useEffect(() => {
        if (err) {
            createErrorNotification(err.message);
        }
    }, [createErrorNotification, err]);

    useEffect(() => { fetch(); }, [fetch]);

    return (
        <LoginLayout
            id="consent-stage"
            title={`Permissions Request`}
            showBrand>
            <Grid container>
                <div>The application {resp?.client_id} would like to</div>
                {resp?.scopes.map(s => <div>{s}</div>)}
                <Button>Accept</Button>
                <Button>Deny</Button>
            </Grid>
        </LoginLayout>
    )
}