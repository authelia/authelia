import React from "react";

import { Delete, Edit, InfoOutlined } from "@mui/icons-material";
import { Box, Paper, Stack, Tooltip, Typography } from "@mui/material";
import IconButton from "@mui/material/IconButton";
import { useTranslation } from "react-i18next";

import { FormatDateHumanReadable } from "@i18n/formats";

interface Props {
    icon?: React.ReactNode;
    description: string;
    qualifier: string;
    created_at: Date;
    last_used_at?: Date;
    tooltipInformation?: string;
    tooltipEdit?: string;
    tooltipDelete: string;
    handleInformation?: (event: React.MouseEvent<HTMLElement>) => void;
    handleEdit?: (event: React.MouseEvent<HTMLElement>) => void;
    handleDelete: (event: React.MouseEvent<HTMLElement>) => void;
}

const CredentialItem = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Paper variant="outlined">
            <Box sx={{ p: 3 }}>
                <Stack direction={"row"} spacing={1} alignItems={"center"}>
                    {props.icon}
                    <Stack spacing={0}>
                        <Box>
                            <Typography display={"inline"} sx={{ fontWeight: "bold" }}>
                                {props.description}
                            </Typography>
                            <Typography display={"inline"} variant={"body2"}>
                                {props.qualifier}
                            </Typography>
                        </Box>
                        <Typography variant={"caption"}>
                            {translate("Added when", {
                                when: props.created_at,
                                formatParams: { when: FormatDateHumanReadable },
                            })}
                        </Typography>
                        <Typography variant={"caption"}>
                            {props.last_used_at === undefined
                                ? translate("Never used")
                                : translate("Last Used when", {
                                      when: props.last_used_at,
                                      formatParams: { when: FormatDateHumanReadable },
                                  })}
                        </Typography>
                    </Stack>
                    {props.handleInformation ? (
                        <TooltipElement tooltip={props.tooltipInformation}>
                            <IconButton color="primary" onClick={props.handleInformation}>
                                <InfoOutlined />
                            </IconButton>
                        </TooltipElement>
                    ) : null}
                    {props.handleEdit ? (
                        <TooltipElement tooltip={props.tooltipEdit}>
                            <IconButton color="primary" onClick={props.handleEdit}>
                                <Edit />
                            </IconButton>
                        </TooltipElement>
                    ) : null}
                    <Tooltip title={props.tooltipDelete}>
                        <IconButton color="primary" onClick={props.handleDelete}>
                            <Delete />
                        </IconButton>
                    </Tooltip>
                </Stack>
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
