import React, { useState } from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import {
    Box,
    CircularProgress,
    Collapse,
    Divider,
    Grid,
    IconButton,
    Stack,
    Switch,
    TableCell,
    TableRow,
    Tooltip,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { WebauthnDevice } from "@root/models/Webauthn";
import { deleteDevice } from "@root/services/Webauthn";

interface Props {
    device: WebauthnDevice;
    webauthnShowDetails: number;
    idx: number;
    handleWebAuthnDetailsChange: (idx: number) => void;
    handleDeleteItem: (idx: number) => void;
}

export default function WebauthnDeviceItem(props: Props) {
    const { t: translate } = useTranslation("settings");
    const { createErrorNotification } = useNotifications();
    const [deleting, setDeleting] = useState(false);

    const handleDelete = async () => {
        setDeleting(true);
        const status = await deleteDevice(props.device.id);
        setDeleting(false);
        if (status !== 200) {
            createErrorNotification(translate("There was a problem deleting the device"));
            return;
        }
        props.handleDeleteItem(props.idx);
    };

    return (
        <React.Fragment>
            <TableRow sx={{ "& > *": { borderBottom: "unset" } }} key={props.device.kid.toString()}>
                <TableCell>
                    <Tooltip title={translate("Show Details")} placement="right">
                        <IconButton
                            aria-label="expand row"
                            size="small"
                            onClick={() => props.handleWebAuthnDetailsChange(props.idx)}
                        >
                            {props.webauthnShowDetails === props.idx ? (
                                <KeyboardArrowUpIcon />
                            ) : (
                                <KeyboardArrowDownIcon />
                            )}
                        </IconButton>
                    </Tooltip>
                </TableCell>
                <TableCell component="th" scope="row">
                    {props.device.description}
                </TableCell>
                <TableCell>
                    <Switch defaultChecked={false} size="small" />
                </TableCell>
                <TableCell align="center">
                    <Stack direction="row" spacing={1} alignItems="center" justifyContent="center">
                        <Tooltip title={translate("Edit")} placement="bottom">
                            <IconButton aria-label="edit">
                                <EditIcon />
                            </IconButton>
                        </Tooltip>
                        {deleting ? (
                            <CircularProgress color="inherit" size={24} />
                        ) : (
                            <Tooltip title={translate("Delete")} placement="bottom">
                                <IconButton aria-label="delete" onClick={handleDelete}>
                                    <DeleteIcon />
                                </IconButton>
                            </Tooltip>
                        )}
                    </Stack>
                </TableCell>
            </TableRow>
            <TableRow>
                <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={4}>
                    <Collapse in={props.webauthnShowDetails === props.idx} timeout="auto" unmountOnExit>
                        <Grid container spacing={2} sx={{ mb: 3, margin: 1 }}>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Box sx={{ margin: 1 }}>
                                    <Typography variant="h6" gutterBottom component="div">
                                        {translate("Details")}
                                    </Typography>
                                </Box>
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Divider variant="middle" />
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Typography>
                                    {translate("Webauthn Credential Identifier", {
                                        id: props.device.kid.toString(),
                                    })}
                                </Typography>
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Typography>
                                    Public Key: {props.device.public_key}
                                    {translate("Webauthn Public Key", {
                                        key: props.device.public_key.toString(),
                                    })}
                                </Typography>
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Divider variant="middle" />
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Relying Party ID")}</Typography>
                                <Typography>{props.device.rpid}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Authenticator Attestation GUID")}</Typography>
                                <Typography>
                                    {props.device.aaguid == undefined ? "N/A" : props.device.aaguid}
                                </Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Attestation Type")}</Typography>
                                <Typography>{props.device.attestation_type}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Transports")}</Typography>
                                <Typography>
                                    {props.device.transports.length === 0 ? "N/A" : props.device.transports.join(", ")}
                                </Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Clone Warning")}</Typography>
                                <Typography>
                                    {props.device.clone_warning ? translate("Yes") : translate("No")}
                                </Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Created")}</Typography>
                                <Typography>{props.device.created_at.toString()}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Last Used")}</Typography>
                                <Typography>
                                    {props.device.last_used_at === undefined
                                        ? translate("Never")
                                        : props.device.last_used_at.toString()}
                                </Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>{translate("Usage Count")}</Typography>
                                <Typography>
                                    {props.device.sign_count === 0 ? translate("Never") : props.device.sign_count}
                                </Typography>
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Divider variant="middle" />
                            </Grid>
                        </Grid>
                    </Collapse>
                </TableCell>
            </TableRow>
        </React.Fragment>
    );
}
