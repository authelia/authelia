import React, { useCallback, useEffect, useRef, useState } from "react";

import {
    Box,
    Button,
    CircularProgress,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Theme,
} from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import { useTranslation } from "react-i18next";

import OneTimeCodeTextField from "@components/OneTimeCodeTextField";
import SuccessIcon from "@components/SuccessIcon";
import { useNotifications } from "@hooks/NotificationsContext";
import {
    UserSessionElevation,
    deleteUserSessionElevation,
    generateUserSessionElevation,
    verifyUserSessionElevation,
} from "@services/UserSessionElevation";

type Props = {
    elevation?: UserSessionElevation;
    opening: boolean;
    handleClosed: (ok: boolean) => void;
    handleOpened: () => void;
};

const IdentityVerificationDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const styles = useStyles();

    const { createErrorNotification } = useNotifications();

    const [open, setOpen] = useState(false);
    const [closing, setClosing] = useState(false);
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);

    const [codeInput, setCodeInput] = useState("");
    const [codeDelete, setCodeDelete] = useState<string>();
    const [codeError, setCodeError] = useState(false);
    const codeRef = useRef<HTMLInputElement>(null);

    const handleClose = useCallback(
        (ok: boolean) => {
            setOpen(false);

            setCodeInput("");
            setCodeDelete(undefined);
            setCodeError(false);
            setLoading(false);
            setSuccess(false);
            setClosing(false);
            props.handleClosed(ok);
        },
        [props],
    );

    const handleDelete = useCallback(async () => {
        if (!codeDelete) {
            throw new Error("The delete code was empty.");
        }

        await deleteUserSessionElevation(codeDelete);
    }, [codeDelete]);

    const handleCancelled = useCallback(() => {
        setClosing(true);

        handleDelete().catch(console.error);

        handleClose(false);
    }, [handleClose, handleDelete]);

    const handleSuccess = useCallback(() => {
        setSuccess(true);

        setTimeout(() => {
            handleClose(true);
        }, 750);
    }, [handleClose]);

    const handleFailure = useCallback(() => {
        setCodeInput("");
        setCodeError(true);
        setLoading(false);

        createErrorNotification(
            translate("The One-Time Code either doesn't match the one generated or an unknown error occurred"),
        );

        codeRef.current?.focus();
    }, [createErrorNotification, translate]);

    const handleLoad = useCallback(async () => {
        if (props.elevation && (props.elevation.elevated || props.elevation.skip_second_factor)) {
            handleClose(true);

            return;
        }

        if (open) {
            return;
        }

        const attempt = await generateUserSessionElevation();

        if (!attempt) throw new Error("Failed to load the data.");

        setCodeDelete(attempt.delete_id);
        props.handleOpened();
        setOpen(true);
    }, [handleClose, open, props]);

    const handleSubmit = useCallback(async () => {
        if (codeInput === "") return;

        setLoading(true);
        const success = await verifyUserSessionElevation(codeInput);

        if (success) {
            handleSuccess();
        } else {
            handleFailure();
        }
    }, [codeInput, handleFailure, handleSuccess]);

    useEffect(() => {
        if (closing || !props.opening || !props.elevation) {
            return;
        }

        handleLoad().catch(console.error);
    }, [closing, handleLoad, props.elevation, props.opening]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        setCodeInput(e.target.value.replace(/\s/g, ""));
        setCodeError(false);
    };

    return (
        <Dialog id={"dialog-verify-one-time-code"} open={open} onClose={handleCancelled}>
            <DialogTitle>{translate("Identity Verification")}</DialogTitle>
            {success ? (
                <DialogContent>
                    <Box
                        className={styles.success}
                        sx={{
                            display: "flex",
                            flexDirection: "column",
                            m: "auto",
                            width: "fit-content",
                            padding: "5.0rem",
                        }}
                    >
                        <SuccessIcon />
                    </Box>
                </DialogContent>
            ) : (
                <DialogContent dividers>
                    <DialogContentText gutterBottom>
                        {translate(
                            "In order to perform this action policy enforcement requires additional identity verification and a One-Time Code has been sent to your email",
                        )}
                    </DialogContentText>
                    <DialogContentText gutterBottom>
                        {translate("Closing this dialog or selecting cancel will invalidate the One-Time Code")}
                    </DialogContentText>
                    <Box
                        sx={{
                            display: "flex",
                            flexDirection: "column",
                            m: "auto",
                            width: "fit-content",
                            marginY: "2.5rem",
                        }}
                    >
                        <OneTimeCodeTextField
                            id={"one-time-code"}
                            label={"One-Time Code"}
                            autoFocus={true}
                            value={codeInput}
                            onChange={handleChange}
                            error={codeError}
                            disabled={loading}
                            inputRef={codeRef}
                        />
                    </Box>
                </DialogContent>
            )}
            {success ? null : (
                <DialogActions>
                    <Button
                        id={"dialog-cancel"}
                        variant={"outlined"}
                        color={"error"}
                        disabled={loading}
                        onClick={handleCancelled}
                    >
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id={"dialog-verify"}
                        variant={"outlined"}
                        color={"primary"}
                        disabled={loading}
                        startIcon={loading ? <CircularProgress color="inherit" size={20} /> : undefined}
                        onClick={handleSubmit}
                    >
                        {translate("Verify")}
                    </Button>
                </DialogActions>
            )}
        </Dialog>
    );
};

export default IdentityVerificationDialog;

const useStyles = makeStyles((theme: Theme) => ({
    centered: {
        display: "flex",
        flexDirection: "column",
        m: "auto",
        width: "fit-content",
        marginY: "2.5rem",
    },
    success: {
        marginBottom: theme.spacing(2),
        flex: "0 0 100%",
        display: "flex",
        flexDirection: "column",
        m: "auto",
        width: "fit-content",
        marginY: "2.5rem",
    },
}));
