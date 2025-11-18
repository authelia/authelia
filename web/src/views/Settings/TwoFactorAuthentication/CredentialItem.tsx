import { MouseEvent, ReactElement, ReactNode } from "react";

import { Delete, Edit, InfoOutlined, ReportProblem } from "@mui/icons-material";
import { Box, Paper, Stack, Tooltip, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";
import IconButton from "@mui/material/IconButton";
import { useTranslation } from "react-i18next";

import { useRelativeTime } from "@hooks/RelativeTimeString";

interface Props {
    id: string;
    icon?: ReactNode;
    description: string;
    qualifier: string;
    problem?: boolean;
    created_at: Date;
    last_used_at?: Date;
    tooltipInformation?: string;
    tooltipInformationProblem?: string;
    tooltipEdit?: string;
    tooltipDelete: string;
    handleInformation?: (_event: MouseEvent<HTMLElement>) => void;
    handleEdit?: (_event: MouseEvent<HTMLElement>) => void;
    handleDelete: (_event: MouseEvent<HTMLElement>) => void;
}

const CredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const timeSinceAdded = useRelativeTime(props.created_at);
    const timeSinceLastUsed = useRelativeTime(props.last_used_at || new Date(0));

    return (
        <Paper variant={"outlined"} id={props.id}>
            <Box sx={{ p: 3 }}>
                <Grid container size={{ xs: 12 }} alignItems={"center"} height={"100%"}>
                    <Grid container size={{ sm: 1, xs: 2 }} marginRight={{ md: 2, xl: 3, xs: 1 }}>
                        {props.icon}
                    </Grid>
                    <Grid size={{ sm: 6, xs: 3 }}>
                        <Stack direction={"column"}>
                            <Stack direction={"row"}>
                                <Typography
                                    id={`${props.id}-description`}
                                    display={"inline"}
                                    sx={{ fontWeight: "bold" }}
                                >
                                    {props.description}
                                </Typography>
                                <Typography display={{ sm: "inline", xs: "none" }} variant={"body2"} px={2}>
                                    {props.qualifier}
                                </Typography>
                            </Stack>
                            <Typography variant={"caption"} display={{ sm: "block", xs: "none" }}>
                                {`${translate("Added")} ${timeSinceAdded}`}
                            </Typography>
                            <Typography variant={"caption"} display={{ sm: "block", xs: "none" }}>
                                {props.last_used_at === undefined
                                    ? translate("Never used")
                                    : `${translate("Last Used")} ${timeSinceLastUsed}`}
                            </Typography>
                        </Stack>
                    </Grid>
                    <Grid size={{ sm: 4, xs: 6 }}>
                        <Grid
                            container
                            size={{ xs: 12 }}
                            justifyContent={"flex-end"}
                            alignItems={"center"}
                            height={"100%"}
                        >
                            {props.handleInformation ? (
                                <Grid size={{ lg: 4, xs: 3 }}>
                                    <TooltipElement
                                        tooltip={
                                            props.problem ? props.tooltipInformationProblem : props.tooltipInformation
                                        }
                                    >
                                        <IconButton
                                            color={"primary"}
                                            onClick={props.handleInformation}
                                            id={`${props.id}-information`}
                                        >
                                            {props.problem ? <ReportProblem color={"warning"} /> : <InfoOutlined />}
                                        </IconButton>
                                    </TooltipElement>
                                </Grid>
                            ) : null}
                            {props.handleEdit ? (
                                <Grid size={{ lg: 4, xs: 3 }}>
                                    <TooltipElement tooltip={props.tooltipEdit}>
                                        <IconButton
                                            color={"primary"}
                                            onClick={props.handleEdit}
                                            id={`${props.id}-edit`}
                                        >
                                            <Edit />
                                        </IconButton>
                                    </TooltipElement>
                                </Grid>
                            ) : null}
                            <Grid size={{ lg: 4, xs: 3 }}>
                                <Tooltip title={props.tooltipDelete}>
                                    <IconButton
                                        color={"primary"}
                                        onClick={props.handleDelete}
                                        id={`${props.id}-delete`}
                                    >
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
    children: ReactElement<any, any>;
}

const TooltipElement = function (props: TooltipElementProps) {
    return props.tooltip !== undefined && props.tooltip !== "" ? (
        <Tooltip title={props.tooltip}>{props.children}</Tooltip>
    ) : (
        props.children
    );
};

export default CredentialItem;
