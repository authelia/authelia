import { ReactNode } from "react";

import { Box, Button, Dialog, DialogActions, DialogContent, Typography, useTheme } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import FingerTouchIcon from "@components/FingerTouchIcon";
import PushNotificationIcon from "@components/PushNotificationIcon";
import TimerIcon from "@components/TimerIcon";
import { SecondFactorMethod } from "@models/Methods";

export interface Props {
    open: boolean;
    methods: Set<SecondFactorMethod>;
    webauthn: boolean;

    onClose: () => void;
    onClick: (_method: SecondFactorMethod) => void;
}

const MethodSelectionDialog = function (props: Props) {
    const { t: translate } = useTranslation();
    const theme = useTheme();
    const pieChartIcon = (
        <TimerIcon width={24} height={24} period={15} color={theme.palette.primary.main} backgroundColor={"white"} />
    );

    return (
        <Dialog open={props.open} sx={{ textAlign: "center" }} onClose={props.onClose}>
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
                <Button color="primary" onClick={props.onClose}>
                    {translate("Close")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

interface MethodItemProps {
    readonly id: string;
    readonly method: string;
    readonly icon: ReactNode;

    readonly onClick: () => void;
}

function MethodItem(props: MethodItemProps) {
    return (
        <Grid size={{ xs: 12 }} className="method-option" id={props.id}>
            <Button
                sx={{
                    display: "block",
                    paddingBottom: (theme) => theme.spacing(4),
                    paddingTop: (theme) => theme.spacing(4),
                    width: "100%",
                }}
                color="primary"
                variant="contained"
                onClick={props.onClick}
            >
                <Box sx={{ display: "inline-block", fill: "white" }}>{props.icon}</Box>
                <Box>
                    <Typography>{props.method}</Typography>
                </Box>
            </Button>
        </Grid>
    );
}

export default MethodSelectionDialog;
