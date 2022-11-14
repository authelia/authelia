import React, { useEffect, useState } from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import {
    Box,
    Button,
    Collapse,
    Divider,
    Grid,
    IconButton,
    Paper,
    Stack,
    Switch,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Tooltip,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { RegisterWebauthnRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { WebauthnDevice } from "@root/models/Webauthn";
import { getWebauthnDevices } from "@root/services/UserWebauthnDevices";
import { initiateWebauthnRegistrationProcess } from "@services/RegisterDevice";
import { AutheliaState, AuthenticationLevel } from "@services/State";

import WebauthnDeviceItem from "./WebauthnDeviceItem";

interface Props {
    state: AutheliaState;
}

export default function TwoFactorAuthSettings(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDevice[] | undefined>();
    const [webauthnShowDetails, setWebauthnShowDetails] = useState<number>(-1);
    const [registrationInProgress, setRegistrationInProgress] = useState(false);

    const handleWebAuthnDetailsChange = (idx: number) => {
        if (webauthnShowDetails === idx) {
            setWebauthnShowDetails(-1);
        } else {
            setWebauthnShowDetails(idx);
        }
    };

    const initiateRegistration = async (initiateRegistrationFunc: () => Promise<void>, redirectRoute: string) => {
        if (props.state.authentication_level >= AuthenticationLevel.TwoFactor) {
            navigate(redirectRoute);
        } else {
            if (registrationInProgress) {
                return;
            }
            setRegistrationInProgress(true);
            try {
                await initiateRegistrationFunc();
                createInfoNotification(translate("An email has been sent to your address to complete the process"));
            } catch (err) {
                console.error(err);
                createErrorNotification(translate("There was a problem initiating the registration process"));
            }
            setRegistrationInProgress(false);
        }
    };

    useEffect(() => {
        (async function () {
            const devices = await getWebauthnDevices();
            setWebauthnDevices(devices);
        })();
    }, []);

    const handleAddKeyButtonClick = () => {
        initiateRegistration(initiateWebauthnRegistrationProcess, RegisterWebauthnRoute);
    };

    return (
        <Grid container spacing={2}>
            <Grid item xs={12}>
                <Typography>{translate("Manage your security keys")}</Typography>
            </Grid>
            <Grid item xs={12}>
                <Stack spacing={1} direction="row">
                    <Button color="primary" variant="contained" onClick={handleAddKeyButtonClick}>
                        {translate("Add")}
                    </Button>
                </Stack>
            </Grid>
            <Grid item xs={12}>
                <Paper>
                    <Table>
                        <TableHead>
                            <TableRow>
                                <TableCell />
                                <TableCell>{translate("Name")}</TableCell>
                                <TableCell>{translate("Enabled")}</TableCell>
                                <TableCell align="center">{translate("Actions")}</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {webauthnDevices
                                ? webauthnDevices.map((x, idx) => {
                                      return (
                                          <WebauthnDeviceItem
                                              device={x}
                                              idx={idx}
                                              webauthnShowDetails={webauthnShowDetails}
                                              handleWebAuthnDetailsChange={handleWebAuthnDetailsChange}
                                          />
                                      );
                                  })
                                : null}
                        </TableBody>
                    </Table>
                </Paper>
            </Grid>
        </Grid>
    );
}
