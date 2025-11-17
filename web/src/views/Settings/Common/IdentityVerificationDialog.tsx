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
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

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
    const { elevation, handleClosed, handleOpened, opening } = props;
    const { t: translate } = useTranslation("settings");
    const { classes } = useStyles();

    const { createErrorNotification } = useNotifications();

    const [closing, setClosing] = useState(false);
    const [loading, setLoading] = useState(false);
    const [success, setSuccess] = useState(false);

    const [codeInput, setCodeInput] = useState("");
    const [codeDelete, setCodeDelete] = useState<string>();
    const [codeError, setCodeError] = useState(false);
    const [ready, setReady] = useState(false);
    const codeRef = useRef<HTMLInputElement>(null);

    const open = React.useMemo(() => ready && !closing && opening && !!elevation, [ready, closing, opening, elevation]);

    const handleClose = useCallback(
        (ok: boolean) => {
            setCodeInput("");
            setCodeDelete(undefined);
            setCodeError(false);
            setLoading(false);
            setSuccess(false);
            setClosing(false);
            setReady(false);
            handleClosed(ok);
        },
        [handleClosed],
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

    const handleSubmitKeyDown = useCallback(
        async (event: React.KeyboardEvent<HTMLDivElement>) => {
            if (event.key === "Enter") {
                if (codeInput.length === 0) {
                    setCodeError(true);
                } else {
                    await handleSubmit();
                }
            }
        },
        [codeInput.length, handleSubmit],
    );

    useEffect(() => {
        if (closing || !opening || !elevation) {
            return;
        }

        if (ready) return;

        generateUserSessionElevation()
            .then((attempt) => {
                if (!attempt) throw new Error("Failed to load the data.");

                setCodeDelete(attempt.delete_id);
                handleOpened();
                setReady(true);
            })
            .catch((error) => {
                console.error(error);
            });
    }, [closing, opening, elevation, ready, handleOpened]);

    const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
        setCodeInput(e.target.value.replaceAll(/\s/g, ""));
        setCodeError(false);
    };

    return (
        <Dialog id={"dialog-verify-one-time-code"} open={open} onClose={handleCancelled}>
            <DialogTitle>{translate("Identity Verification")}</DialogTitle>
            {success ? (
                <DialogContent>
                    <Box
                        className={classes.success}
                        sx={{
                            display: "flex",
                            flexDirection: "column",
                            m: "auto",
                            padding: "5.0rem",
                            width: "fit-content",
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
                            marginY: "2.5rem",
                            width: "fit-content",
                        }}
                    >
                        <OneTimeCodeTextField
                            id={"one-time-code"}
                            label={"One-Time Code"}
                            value={codeInput}
                            onChange={handleChange}
                            error={codeError}
                            disabled={loading}
                            inputRef={codeRef}
                            onKeyDown={handleSubmitKeyDown}
                        />
                    </Box>
                </DialogContent>
            )}
            {success ? null : (
                <DialogActions>
                    <Button
                        id={"dialog-cancel"}
                        variant={"contained"}
                        color={"error"}
                        disabled={loading}
                        onClick={handleCancelled}
                    >
                        {translate("Cancel")}
                    </Button>
                    <Button
                        id={"dialog-verify"}
                        variant={"contained"}
                        color={"info"}
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

const useStyles = makeStyles()((theme: Theme) => ({
    success: {
        display: "flex",
        flex: "0 0 100%",
        flexDirection: "column",
        m: "auto",
        marginBottom: theme.spacing(2),
        marginY: "2.5rem",
        width: "fit-content",
    },
}));

export default IdentityVerificationDialog;
