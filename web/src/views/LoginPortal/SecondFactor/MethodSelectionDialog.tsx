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
                    {props.methods.has(SecondFactorMethod.Telegram) ? (
                        <MethodItem
                            id="telegram-option"
                            method={translate("Telegram")}
                            icon={
                                <svg width="32" height="32" viewBox="0 0 24 24" fill="currentColor">
                                    <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69a.26.26 0 00-.07-.2c-.08-.06-.2-.04-.28-.02-.12.03-1.98 1.26-5.61 3.71-.53.37-1.01.55-1.44.54-.47-.01-1.38-.27-2.06-.49-.83-.27-1.49-.42-1.43-.88.03-.24.37-.49 1.02-.75 3.97-1.73 6.62-2.87 7.94-3.44 3.78-1.57 4.57-1.85 5.08-1.86.11 0 .37.03.54.17.14.12.18.28.2.47 0 .06.01.24 0 .38z" />
                                </svg>
                            }
                            onClick={() => props.onClick(SecondFactorMethod.Telegram)}
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
