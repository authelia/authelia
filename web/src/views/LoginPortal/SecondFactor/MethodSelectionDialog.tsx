import React, { ReactNode } from "react";

import {
    Dialog,
    Grid,
    makeStyles,
    DialogContent,
    Button,
    DialogActions,
    Typography,
    useTheme,
} from "@material-ui/core";

import FingerTouchIcon from "../../../components/FingerTouchIcon";
import PieChartIcon from "../../../components/PieChartIcon";
import PushNotificationIcon from "../../../components/PushNotificationIcon";
import { SecondFactorMethod } from "../../../models/Methods";

export interface Props {
    open: boolean;
    methods: Set<SecondFactorMethod>;
    u2fSupported: boolean;

    onClose: () => void;
    onClick: (method: SecondFactorMethod) => void;
}

const MethodSelectionDialog = function (props: Props) {
    const style = useStyles();
    const theme = useTheme();

    const pieChartIcon = (
        <PieChartIcon
            width={24}
            height={24}
            maxProgress={1000}
            progress={150}
            color={theme.palette.primary.main}
            backgroundColor={"white"}
        />
    );

    return (
        <Dialog open={props.open} className={style.root} onClose={props.onClose}>
            <DialogContent>
                <Grid container justify="center" spacing={1} id="methods-dialog">
                    {props.methods.has(SecondFactorMethod.TOTP) ? (
                        <MethodItem
                            id="one-time-password-option"
                            method="One-Time Password"
                            icon={pieChartIcon}
                            onClick={() => props.onClick(SecondFactorMethod.TOTP)}
                        />
                    ) : null}
                    {props.methods.has(SecondFactorMethod.U2F) && props.u2fSupported ? (
                        <MethodItem
                            id="security-key-option"
                            method="Security Key"
                            icon={<FingerTouchIcon size={32} />}
                            onClick={() => props.onClick(SecondFactorMethod.U2F)}
                        />
                    ) : null}
                    {props.methods.has(SecondFactorMethod.MobilePush) ? (
                        <MethodItem
                            id="push-notification-option"
                            method="Push Notification"
                            icon={<PushNotificationIcon width={32} height={32} />}
                            onClick={() => props.onClick(SecondFactorMethod.MobilePush)}
                        />
                    ) : null}
                </Grid>
            </DialogContent>
            <DialogActions>
                <Button color="primary" onClick={props.onClose}>
                    Close
                </Button>
            </DialogActions>
        </Dialog>
    );
};

export default MethodSelectionDialog;

const useStyles = makeStyles((theme) => ({
    root: {
        textAlign: "center",
    },
}));

interface MethodItemProps {
    id: string;
    method: string;
    icon: ReactNode;

    onClick: () => void;
}

function MethodItem(props: MethodItemProps) {
    const style = makeStyles((theme) => ({
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
    }))();

    return (
        <Grid item xs={12} className="method-option" id={props.id}>
            <Button
                className={style.item}
                color="primary"
                classes={{ root: style.buttonRoot }}
                variant="contained"
                onClick={props.onClick}
            >
                <div className={style.icon}>{props.icon}</div>
                <div>
                    <Typography>{props.method}</Typography>
                </div>
            </Button>
        </Grid>
    );
}
