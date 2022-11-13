import React, { useEffect, useState } from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import SystemSecurityUpdateGoodIcon from "@mui/icons-material/SystemSecurityUpdateGood";
import {
    AppBar,
    Box,
    Button,
    Collapse,
    Divider,
    Drawer,
    Grid,
    IconButton,
    List,
    ListItem,
    ListItemButton,
    ListItemIcon,
    ListItemText,
    Paper,
    Stack,
    Switch,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Toolbar,
    Tooltip,
    Typography,
} from "@mui/material";

import { WebauthnDevice } from "@root/models/Webauthn";
import { getWebauthnDevices } from "@root/services/UserWebauthnDevices";

import AddSecurityKeyDialog from "./AddSecurityDialog";

interface Props {}

const drawerWidth = 240;

export default function SettingsView(props: Props) {
    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDevice[] | undefined>();
    const [addKeyOpen, setAddKeyOpen] = useState<boolean>(false);

    useEffect(() => {
        (async function () {
            const devices = await getWebauthnDevices();
            setWebauthnDevices(devices);
        })();
    }, []);

    const handleKeyClose = () => {
        setAddKeyOpen(false);
    };

    const handleAddKeyButtonClick = () => {
        setAddKeyOpen(true);
    };

    return (
        <Box sx={{ display: "flex" }}>
            <AppBar position="fixed" sx={{ zIndex: (theme) => theme.zIndex.drawer + 1 }}>
                <Toolbar variant="dense">
                    <Typography style={{ flexGrow: 1 }}>Settings</Typography>
                </Toolbar>
            </AppBar>
            <Drawer
                variant="permanent"
                sx={{
                    width: drawerWidth,
                    flexShrink: 0,
                    [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: "border-box" },
                }}
            >
                <Toolbar variant="dense" />
                <Box sx={{ overflow: "auto" }}>
                    <List>
                        <ListItem disablePadding>
                            <ListItemButton selected={true}>
                                <ListItemIcon>
                                    <SystemSecurityUpdateGoodIcon />
                                </ListItemIcon>
                                <ListItemText primary={"Security Keys"} />
                            </ListItemButton>
                        </ListItem>
                    </List>
                </Box>
            </Drawer>
            <Box component="main" sx={{ flexGrow: 1, p: 3 }}>
                <Grid container spacing={2}>
                    <Grid item xs={12}>
                        <Typography>Manage your security keys</Typography>
                    </Grid>
                    <Grid item xs={12}>
                        <Stack spacing={1} direction="row">
                            <Button color="primary" variant="contained" onClick={handleAddKeyButtonClick}>
                                Add
                            </Button>
                        </Stack>
                    </Grid>
                    <Grid item xs={12}>
                        <Paper>
                            <Table>
                                <TableHead>
                                    <TableRow>
                                        <TableCell />
                                        <TableCell>Name</TableCell>
                                        <TableCell>Enabled</TableCell>
                                        <TableCell align="center">Actions</TableCell>
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {webauthnDevices
                                        ? webauthnDevices.map((x, idx) => {
                                              return <WebauthnDeviceRow device={x} />;
                                          })
                                        : null}
                                </TableBody>
                            </Table>
                        </Paper>
                    </Grid>
                </Grid>
            </Box>
            <AddSecurityKeyDialog open={addKeyOpen} onClose={handleKeyClose} />
        </Box>
    );
}

interface WebauthnDeviceRowProps {
    device: WebauthnDevice;
}

function WebauthnDeviceRow(props: WebauthnDeviceRowProps) {
    const [showDetails, setShowDetails] = useState<boolean>(false);

    return (
        <React.Fragment>
            <TableRow sx={{ "& > *": { borderBottom: "unset" } }} key={props.device.kid.toString()}>
                <TableCell>
                    <Tooltip title="Show Details" placement="right">
                        <IconButton aria-label="expand row" size="small" onClick={() => setShowDetails(!showDetails)}>
                            {showDetails ? <KeyboardArrowUpIcon /> : <KeyboardArrowDownIcon />}
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
                        <Tooltip title="Edit" placement="bottom">
                            <IconButton aria-label="edit">
                                <EditIcon />
                            </IconButton>
                        </Tooltip>
                        <Tooltip title="Delete" placement="bottom">
                            <IconButton aria-label="delete">
                                <DeleteIcon />
                            </IconButton>
                        </Tooltip>
                    </Stack>
                </TableCell>
            </TableRow>
            <TableRow>
                <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={4}>
                    <Collapse in={showDetails} timeout="auto" unmountOnExit>
                        <Grid container spacing={2} sx={{ mb: 3, margin: 1 }}>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Box sx={{ margin: 1 }}>
                                    <Typography variant="h6" gutterBottom component="div">
                                        Details
                                    </Typography>
                                </Box>
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Divider variant="middle" />
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Typography>Key ID: {props.device.kid}</Typography>
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Typography>Public Key: {props.device.public_key}</Typography>
                            </Grid>
                            <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                <Divider variant="middle" />
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Relying Party ID</Typography>
                                <Typography>{props.device.rpid}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Authenticator Attestation GUID</Typography>
                                <Typography>{props.device.aaguid}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Attestation Type</Typography>
                                <Typography>{props.device.attestation_type}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Transports</Typography>
                                <Typography>
                                    {props.device.transports.length === 0 ? "N/A" : props.device.transports.join(", ")}
                                </Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Clone Warning</Typography>
                                <Typography>{props.device.clone_warning ? "Yes" : "No"}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Created</Typography>
                                <Typography>{props.device.created_at.toString()}</Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Last Used</Typography>
                                <Typography>
                                    {props.device.last_used_at === undefined
                                        ? "Never"
                                        : props.device.last_used_at.toString()}
                                </Typography>
                            </Grid>
                            <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                <Typography>Usage Count</Typography>
                                <Typography>
                                    {props.device.sign_count === 0 ? "Never" : props.device.sign_count}
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
