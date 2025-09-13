import React, { ReactNode } from "react";

import { Box, Button, Dialog, DialogActions, DialogContent, Theme, Typography, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

import FingerTouchIcon from "@components/FingerTouchIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import TimerIcon from "@components/TimerIcon";
import { SecondFactorMethod } from "@models/Methods";

export interface Props {
    open: boolean;
    methods: Set<SecondFactorMethod>;
    webauthn: boolean;

    onClose: () => void;
    onClick: (method: SecondFactorMethod) => void;
}

const MethodSelectionDialog = function (props: Props) {
    const { t: translate } = useTranslation();
    const theme = useTheme();
    const { classes } = useStyles();

    const pieChartIcon = (
        <TimerIcon width={24} height={24} period={15} color={theme.palette.primary.main} backgroundColor={"white"} />
    );

    return (
        <Dialog open={props.open} className={classes.root} onClose={props.onClose}>
            <DialogContent>
                <Grid container justifyContent="center" spacing={1} id="methods-dialog">
                    {props.methods.has(SecondFactorMethod.TOTP) ? (
                        <MethodItem
                            id="one-time-password-option"
                            method={translate("Time-based One-Time Password")}
                            icon={pieChartIcon}
                            onClick={() => props.onClick(SecondFactorMethod.TOTP)}
                        />
                    ) : null}
                    {props.methods.has(SecondFactorMethod.WebAuthn) && props.webauthn ? (
                        <MethodItem
                            id="webauthn-option"
                            method={translate("Security Key - WebAuthn")}
                            icon={<FingerTouchIcon size={32} />}
                            onClick={() => props.onClick(SecondFactorMethod.WebAuthn)}
                        />
                    ) : null}
                    {props.methods.has(SecondFactorMethod.MobilePush) ? (
                        <MethodItem
                            id="push-notification-option"
                            method={translate("Push Notification")}
                            icon={<PushNotificationIcon width={32} height={32} />}
                            onClick={() => props.onClick(SecondFactorMethod.MobilePush)}
                        />
                    ) : null}
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button color="primary" onClick={props.onClose} data-1p-ignore>
                    {translate("Close")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

interface MethodItemProps {
    id: string;
    method: string;
    icon: ReactNode;

    onClick: () => void;
}

function MethodItem(props: MethodItemProps) {
    const { classes } = useStyles();

    return (
        <Grid size={{ xs: 12 }} className="method-option" id={props.id}>
            <Button
                className={classes.item}
                color="primary"
                classes={{ root: classes.buttonRoot }}
                variant="contained"
                onClick={props.onClick}
                data-1p-ignore
            >
                <Box className={classes.icon}>{props.icon}</Box>
                <Box>
                    <Typography>{props.method}</Typography>
                </Box>
            </Button>
        </Grid>
    );
}

const useStyles = makeStyles()((theme: Theme) => ({
    root: {
        textAlign: "center",
    },
    item: {
        paddingTop: theme.spacing(4),
        paddingBottom: theme.spacing(4),
        width: "100%",
    },
    icon: {
        display: "inline-block",
        fill: "white",
    },
    buttonRoot: {
        display: "block",
    },
}));

export default MethodSelectionDialog;
