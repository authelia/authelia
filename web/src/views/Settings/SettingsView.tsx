import React, { useEffect, useState } from "react";

import { AppBar, Box, Button, Drawer, Grid, IconButton, List, ListItem, ListItemButton, ListItemIcon, ListItemText, Paper, Stack, Switch, Table, TableBody, TableCell, TableHead, TableRow, Toolbar, Tooltip, Typography } from "@mui/material";
import SystemSecurityUpdateGoodIcon from '@mui/icons-material/SystemSecurityUpdateGood';
import DeleteIcon from '@mui/icons-material/Delete';
import EditIcon from '@mui/icons-material/Edit';
import { getWebauthnDevices } from "@root/services/UserWebauthnDevices";
import { WebauthnDevice } from "@root/models/Webauthn";
import AddSecurityKeyDialog from "./AddSecurityDialog";

interface Props {}

const drawerWidth = 240;

export default function SettingsView(props: Props) {
    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDevice[] | undefined>();
    const [addKeyOpen, setAddKeyOpen] = useState<boolean>(false);

    useEffect(() => {
        (async function() {
            const devices = await getWebauthnDevices();
            setWebauthnDevices(devices);
        })()
    }, []);

    const handleKeyClose = () => {
        setAddKeyOpen(false);
    }

    const handleAddKeyButtonClick = () => {
        setAddKeyOpen(true);
    }

    return (
        <Box sx={{ display: 'flex' }}>
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
                [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: 'border-box' },
                }}
            >
                <Toolbar variant="dense" />
                <Box sx={{ overflow: 'auto' }}>
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
                            <Button color="primary" variant="contained" onClick={handleAddKeyButtonClick}>Add</Button>
                        </Stack>
                    </Grid>
                    <Grid item xs={12}>
                        <Paper>
                            <Table>
                                <TableHead>
                                    <TableRow>
                                        <TableCell>Name</TableCell>
                                        <TableCell>Enabled</TableCell>
                                        <TableCell>Activation</TableCell>
                                        <TableCell>Public Key</TableCell>
                                        <TableCell>Actions</TableCell> 
                                    </TableRow>
                                </TableHead>
                                <TableBody>
                                    {webauthnDevices ? webauthnDevices.map((x, idx) => {
                                        return (
                                            <TableRow key={x.description}>
                                                <TableCell>{x.description}</TableCell>
                                                <TableCell><Switch defaultChecked={false} size="small" /></TableCell>
                                                <TableCell><Typography>{(false) ? "<ADATE>" : "Not enabled"}</Typography></TableCell>
                                                <TableCell>
                                                    <Tooltip title={x.public_key}>
                                                    <div style={{overflow: "hidden", textOverflow: "ellipsis", width: '300px'}}>
                                                        <Typography noWrap>{x.public_key}</Typography>
                                                    </div>
                                                    </Tooltip>
                                                </TableCell>
                                                <TableCell>
                                                    <Stack direction="row" spacing={1}>
                                                        <IconButton aria-label="edit">
                                                            <EditIcon />
                                                        </IconButton>
                                                        <IconButton aria-label="delete">
                                                            <DeleteIcon />
                                                        </IconButton>
                                                    </Stack>
                                                </TableCell>
                                            </TableRow>
                                        ) 
                                    }) : null}
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
