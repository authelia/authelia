import React, { useEffect, useState } from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import KeyboardArrowDownIcon from "@mui/icons-material/KeyboardArrowDown";
import KeyboardArrowUpIcon from "@mui/icons-material/KeyboardArrowUp";
import {
    Box,
    Button,
    Collapse,
    Divider,
    Grid,
    IconButton,
    Paper,
    Stack,
    Switch,
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableRow,
    Tooltip,
    Typography,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

import { RegisterWebauthnRoute } from "@constants/Routes";
import { useNotifications } from "@hooks/NotificationsContext";
import { WebauthnDevice } from "@root/models/Webauthn";
import { getWebauthnDevices } from "@root/services/UserWebauthnDevices";
import { initiateWebauthnRegistrationProcess } from "@services/RegisterDevice";
import { AutheliaState, AuthenticationLevel } from "@services/State";

interface Props {
    state: AutheliaState;
}

export default function TwoFactorAuthSettings(props: Props) {
    const { t: translate } = useTranslation("settings");
    const navigate = useNavigate();

    const { createInfoNotification, createErrorNotification } = useNotifications();
    const [webauthnDevices, setWebauthnDevices] = useState<WebauthnDevice[] | undefined>();
    const [webauthnShowDetails, setWebauthnShowDetails] = useState<number>(-1);
    const [registrationInProgress, setRegistrationInProgress] = useState(false);

    const handleWebAuthnDetailsChange = (idx: number) => {
        if (webauthnShowDetails === idx) {
            setWebauthnShowDetails(-1);
        } else {
            setWebauthnShowDetails(idx);
        }
    };

    const initiateRegistration = async (initiateRegistrationFunc: () => Promise<void>, redirectRoute: string) => {
        if (props.state.authentication_level >= AuthenticationLevel.TwoFactor) {
            navigate(redirectRoute);
        } else {
            if (registrationInProgress) {
                return;
            }
            setRegistrationInProgress(true);
            try {
                await initiateRegistrationFunc();
                createInfoNotification(translate("An email has been sent to your address to complete the process"));
            } catch (err) {
                console.error(err);
                createErrorNotification(translate("There was a problem initiating the registration process"));
            }
            setRegistrationInProgress(false);
        }
    };

    useEffect(() => {
        (async function () {
            const devices = await getWebauthnDevices();
            setWebauthnDevices(devices);
        })();
    }, []);

    const handleAddKeyButtonClick = () => {
        initiateRegistration(initiateWebauthnRegistrationProcess, RegisterWebauthnRoute);
    };

    return (
        <Grid container spacing={2}>
            <Grid item xs={12}>
                <Typography>{translate("Manage your security keys")}</Typography>
            </Grid>
            <Grid item xs={12}>
                <Stack spacing={1} direction="row">
                    <Button color="primary" variant="contained" onClick={handleAddKeyButtonClick}>
                        {translate("Add")}
                    </Button>
                </Stack>
            </Grid>
            <Grid item xs={12}>
                <Paper>
                    <Table>
                        <TableHead>
                            <TableRow>
                                <TableCell />
                                <TableCell>{translate("Name")}</TableCell>
                                <TableCell>{translate("Enabled")}</TableCell>
                                <TableCell align="center">{translate("Actions")}</TableCell>
                            </TableRow>
                        </TableHead>
                        <TableBody>
                            {webauthnDevices
                                ? webauthnDevices.map((x, idx) => {
                                      return (
                                          <React.Fragment>
                                              <TableRow
                                                  sx={{ "& > *": { borderBottom: "unset" } }}
                                                  key={x.kid.toString()}
                                              >
                                                  <TableCell>
                                                      <Tooltip title={translate("Show Details")} placement="right">
                                                          <IconButton
                                                              aria-label="expand row"
                                                              size="small"
                                                              onClick={() => handleWebAuthnDetailsChange(idx)}
                                                          >
                                                              {webauthnShowDetails === idx ? (
                                                                  <KeyboardArrowUpIcon />
                                                              ) : (
                                                                  <KeyboardArrowDownIcon />
                                                              )}
                                                          </IconButton>
                                                      </Tooltip>
                                                  </TableCell>
                                                  <TableCell component="th" scope="row">
                                                      {x.description}
                                                  </TableCell>
                                                  <TableCell>
                                                      <Switch defaultChecked={false} size="small" />
                                                  </TableCell>
                                                  <TableCell align="center">
                                                      <Stack
                                                          direction="row"
                                                          spacing={1}
                                                          alignItems="center"
                                                          justifyContent="center"
                                                      >
                                                          <Tooltip title={translate("Edit")} placement="bottom">
                                                              <IconButton aria-label="edit">
                                                                  <EditIcon />
                                                              </IconButton>
                                                          </Tooltip>
                                                          <Tooltip title={translate("Delete")} placement="bottom">
                                                              <IconButton aria-label="delete">
                                                                  <DeleteIcon />
                                                              </IconButton>
                                                          </Tooltip>
                                                      </Stack>
                                                  </TableCell>
                                              </TableRow>
                                              <TableRow>
                                                  <TableCell style={{ paddingBottom: 0, paddingTop: 0 }} colSpan={4}>
                                                      <Collapse
                                                          in={webauthnShowDetails === idx}
                                                          timeout="auto"
                                                          unmountOnExit
                                                      >
                                                          <Grid container spacing={2} sx={{ mb: 3, margin: 1 }}>
                                                              <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                                                  <Box sx={{ margin: 1 }}>
                                                                      <Typography
                                                                          variant="h6"
                                                                          gutterBottom
                                                                          component="div"
                                                                      >
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
                                                                          id: x.kid.toString(),
                                                                      })}
                                                                  </Typography>
                                                              </Grid>
                                                              <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                                                  <Typography>
                                                                      Public Key: {x.public_key}
                                                                      {translate("Webauthn Public Key", {
                                                                          key: x.public_key.toString(),
                                                                      })}
                                                                  </Typography>
                                                              </Grid>
                                                              <Grid item xs={12} sm={12} md={12} lg={12} xl={12}>
                                                                  <Divider variant="middle" />
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>
                                                                      {translate("Relying Party ID")}
                                                                  </Typography>
                                                                  <Typography>{x.rpid}</Typography>
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>
                                                                      {translate("Authenticator Attestation GUID")}
                                                                  </Typography>
                                                                  <Typography>{x.aaguid}</Typography>
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>
                                                                      {translate("Attestation Type")}
                                                                  </Typography>
                                                                  <Typography>{x.attestation_type}</Typography>
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>{translate("Transports")}</Typography>
                                                                  <Typography>
                                                                      {x.transports.length === 0
                                                                          ? "N/A"
                                                                          : x.transports.join(", ")}
                                                                  </Typography>
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>{translate("Clone Warning")}</Typography>
                                                                  <Typography>
                                                                      {x.clone_warning
                                                                          ? translate("Yes")
                                                                          : translate("No")}
                                                                  </Typography>
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>{translate("Created")}</Typography>
                                                                  <Typography>{x.created_at.toString()}</Typography>
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>{translate("Last Used")}</Typography>
                                                                  <Typography>
                                                                      {x.last_used_at === undefined
                                                                          ? translate("Never")
                                                                          : x.last_used_at.toString()}
                                                                  </Typography>
                                                              </Grid>
                                                              <Grid item xs={6} sm={6} md={4} lg={4} xl={3}>
                                                                  <Typography>{translate("Usage Count")}</Typography>
                                                                  <Typography>
                                                                      {x.sign_count === 0
                                                                          ? translate("Never")
                                                                          : x.sign_count}
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
                                  })
                                : null}
                        </TableBody>
                    </Table>
                </Paper>
            </Grid>
        </Grid>
    );
}
