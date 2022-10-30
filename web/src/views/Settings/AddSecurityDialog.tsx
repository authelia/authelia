import { Button, Dialog, DialogActions, DialogContent, DialogContentText, DialogProps, DialogTitle, TextField } from "@mui/material";
import React from "react";

interface Props extends DialogProps {};

export default function AddSecurityKeyDialog(props: Props) {
    const handleAddClick = () => {
        if (props.onClose) {
            props.onClose({}, "backdropClick");
        }
    }

    const handleCancelClick = () => {
        if (props.onClose) {
            props.onClose({}, "backdropClick");
        }
    }

    return (
        <Dialog {...props}>
            <DialogTitle>Add new Security Key</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Provide the details for the new security key.
                </DialogContentText>
                <TextField
                    autoFocus
                    margin="dense"
                    id="description"
                    label="Description"
                    type="text"
                    fullWidth
                    variant="standard"
                />
            </DialogContent>
            <DialogActions>
                <Button onClick={handleCancelClick}>Cancel</Button>
                <Button onClick={handleAddClick}>Add</Button>
            </DialogActions>
        </Dialog>
    );
}