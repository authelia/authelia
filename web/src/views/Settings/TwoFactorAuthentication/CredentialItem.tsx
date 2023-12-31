import React from "react";

import { Delete, Edit, InfoOutlined, ReportProblem } from "@mui/icons-material";
import { Box, Paper, Stack, Tooltip, Typography } from "@mui/material";
import IconButton from "@mui/material/IconButton";
import Grid from "@mui/material/Unstable_Grid2/Grid2";
import { useTranslation } from "react-i18next";

import { FormatDateHumanReadable } from "@i18n/formats";

interface Props {
    id: string;
    icon?: React.ReactNode;
    description: string;
    qualifier: string;
    problem?: boolean;
    created_at: Date;
    last_used_at?: Date;
    tooltipInformation?: string;
    tooltipInformationProblem?: string;
    tooltipEdit?: string;
    tooltipDelete: string;
    handleInformation?: (event: React.MouseEvent<HTMLElement>) => void;
    handleEdit?: (event: React.MouseEvent<HTMLElement>) => void;
    handleDelete: (event: React.MouseEvent<HTMLElement>) => void;
}

const CredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Paper variant="outlined" id={props.id}>
            <Box sx={{ p: 3 }}>
                <Grid container xs={12} alignItems={"center"} height={"100%"}>
                    <Grid xs={2} sm={1} marginRight={{ xs: 1, md: 2, xl: 3 }}>
                        {props.icon}
                    </Grid>
                    <Grid xs={3} sm={6}>
                        <Stack direction={"column"}>
                            <Stack direction={"row"}>
                                <Typography
                                    id={`${props.id}-description`}
                                    display={"inline"}
                                    sx={{ fontWeight: "bold" }}
                                >
                                    {props.description}
                                </Typography>
                                <Typography display={{ xs: "none", sm: "inline" }} variant={"body2"} px={2}>
                                    {props.qualifier}
                                </Typography>
                            </Stack>
                            <Typography variant={"caption"} display={{ xs: "none", sm: "block" }}>
                                {translate("Added when", {
                                    when: props.created_at,
                                    formatParams: { when: FormatDateHumanReadable },
                                })}
                            </Typography>
                            <Typography variant={"caption"} display={{ xs: "none", sm: "block" }}>
                                {props.last_used_at === undefined
                                    ? translate("Never used")
                                    : translate("Last Used when", {
                                          when: props.last_used_at,
                                          formatParams: { when: FormatDateHumanReadable },
                                      })}
                            </Typography>
                        </Stack>
                    </Grid>
                    <Grid xs={6} sm={4}>
                        <Grid container xs={12} justifyContent={"flex-end"} alignItems={"center"} height={"100%"}>
                            {props.handleInformation ? (
                                <Grid xs={2} sm={1}>
                                    <TooltipElement
                                        tooltip={
                                            props.problem ? props.tooltipInformationProblem : props.tooltipInformation
                                        }
                                    >
                                        <IconButton
                                            color="primary"
                                            onClick={props.handleInformation}
                                            id={`${props.id}-information`}
                                        >
                                            {props.problem ? <ReportProblem color={"warning"} /> : <InfoOutlined />}
                                        </IconButton>
                                    </TooltipElement>
                                </Grid>
                            ) : null}
                            {props.handleEdit ? (
                                <Grid xs={2} sm={1}>
                                    <TooltipElement tooltip={props.tooltipEdit}>
                                        <IconButton color="primary" onClick={props.handleEdit} id={`${props.id}-edit`}>
                                            <Edit />
                                        </IconButton>
                                    </TooltipElement>
                                </Grid>
                            ) : null}
                            <Grid xs={2} sm={1}>
                                <Tooltip title={props.tooltipDelete}>
                                    <IconButton color="primary" onClick={props.handleDelete} id={`${props.id}-delete`}>
                                        <Delete />
                                    </IconButton>
                                </Tooltip>
                            </Grid>
                        </Grid>
                    </Grid>
                </Grid>
            </Box>
        </Paper>
    );
};

interface TooltipElementProps {
    tooltip?: string;
    children: React.ReactElement<any, any>;
}

const TooltipElement = function (props: TooltipElementProps) {
    return props.tooltip !== undefined && props.tooltip !== "" ? (
        <Tooltip title={props.tooltip}>{props.children}</Tooltip>
    ) : (
        props.children
    );
};

export default CredentialItem;
