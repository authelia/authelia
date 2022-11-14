import React from "react";

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogProps,
    DialogTitle,
    TextField,
} from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props extends DialogProps {}

export default function AddSecurityKeyDialog(props: Props) {
    const { t: translate } = useTranslation("settings");

    const handleAddClick = () => {
        if (props.onClose) {
            props.onClose({}, "backdropClick");
        }
    };

    const handleCancelClick = () => {
        if (props.onClose) {
            props.onClose({}, "backdropClick");
        }
    };

    return (
        <Dialog {...props}>
            <DialogTitle>{translate("Add new Security Key")}</DialogTitle>
            <DialogContent>
                <DialogContentText>{translate("Provide the details for the new security key")}.</DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="description"
                    label={translate("Description")}
                    type="text"
                    fullWidth
                    variant="standard"
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancelClick}>{translate("Cancel")}</Button>
                <Button onClick={handleAddClick}>{translate("Add")}</Button>
            </DialogActions>
        </Dialog>
    );
}
