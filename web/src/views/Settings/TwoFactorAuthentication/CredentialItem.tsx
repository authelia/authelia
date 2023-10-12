import React from "react";

import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import InfoOutlinedIcon from "@mui/icons-material/InfoOutlined";
import { Paper, Stack, Tooltip, Typography } from "@mui/material";
import IconButton from "@mui/material/IconButton";
import Grid from "@mui/material/Unstable_Grid2";
import { useTranslation } from "react-i18next";

import { FormatDateHumanReadable } from "@i18n/formats";

interface Props {
    icon: React.ReactNode;
    description: string;
    created_at: Date;
    last_used_at?: Date;
    tooltipInformation?: string;
    tooltipEdit?: string;
    tooltipDelete?: string;
    handleDelete: (event: React.MouseEvent<HTMLElement>) => void;
    handleInformation?: (event: React.MouseEvent<HTMLElement>) => void;
    handleEdit?: (event: React.MouseEvent<HTMLElement>) => void;
}

const CredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Grid xs={12} md={6} xl={3}>
            <Paper variant="outlined">
                <Grid container spacing={1} alignItems="center" padding={3}>
                    <Grid xs={12} sm={6} md={6}>
                        <Grid container>
                            <Grid xs={12}>
                                <Stack direction={"row"} spacing={1} alignItems={"center"}>
                                    {props.icon}
                                    <Typography display="inline" sx={{ fontWeight: "bold" }}>
                                        {props.description}
                                    </Typography>
                                    <Typography display="inline" variant="body2">{``}</Typography>
                                </Stack>
                            </Grid>
                            <Grid xs={12} sx={{ display: { xs: "none", md: "block" } }}>
                                <Stack direction={"row"} spacing={1} alignItems={"center"}>
                                    <Typography variant={"caption"} sx={{ display: { xs: "none", md: "block" } }}>
                                        {translate("Added when", {
                                            when: props.created_at,
                                            formatParams: { when: FormatDateHumanReadable },
                                        })}
                                    </Typography>
                                    <Typography variant={"caption"} sx={{ display: { xs: "none", md: "block" } }}>
                                        {props.last_used_at === undefined
                                            ? translate("Never used")
                                            : translate("Last Used when", {
                                                  when: props.last_used_at,
                                                  formatParams: { when: FormatDateHumanReadable },
                                              })}
                                    </Typography>
                                </Stack>
                            </Grid>
                        </Grid>
                    </Grid>
                    <Grid xs={12} md={7} xl={5}>
                        <Stack direction={"row"} spacing={1}>
                            {props.handleInformation ? (
                                props.tooltipInformation ? (
                                    <Tooltip title={props.tooltipInformation}>
                                        <IconButton color="primary" onClick={props.handleInformation}>
                                            <InfoOutlinedIcon />
                                        </IconButton>
                                    </Tooltip>
                                ) : (
                                    <IconButton color="primary" onClick={props.handleInformation}>
                                        <DeleteIcon />
                                    </IconButton>
                                )
                            ) : null}
                            {props.handleInformation ? (
                                props.tooltipEdit ? (
                                    <Tooltip title={props.tooltipEdit}>
                                        <IconButton color="primary" onClick={props.handleEdit}>
                                            <EditIcon />
                                        </IconButton>
                                    </Tooltip>
                                ) : (
                                    <IconButton color="primary" onClick={props.handleEdit}>
                                        <DeleteIcon />
                                    </IconButton>
                                )
                            ) : null}
                            {props.tooltipDelete ? (
                                <Tooltip title={props.tooltipDelete}>
                                    <IconButton color="primary" onClick={props.handleDelete}>
                                        <DeleteIcon />
                                    </IconButton>
                                </Tooltip>
                            ) : (
                                <IconButton color="primary" onClick={props.handleDelete}>
                                    <DeleteIcon />
                                </IconButton>
                            )}
                        </Stack>
                    </Grid>
                </Grid>
            </Paper>
        </Grid>
    );
};

export default CredentialItem;
