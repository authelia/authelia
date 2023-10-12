import React, { Fragment } from "react";

import { QrCode2 } from "@mui/icons-material";
import DeleteIcon from "@mui/icons-material/Delete";
import { Box, Paper, Stack, Tooltip, Typography } from "@mui/material";
import IconButton from "@mui/material/IconButton";
import { useTranslation } from "react-i18next";

import { FormatDateHumanReadable } from "@i18n/formats.ts";
import { UserInfoTOTPConfiguration, toAlgorithmString } from "@models/TOTPConfiguration";

interface Props {
    config: UserInfoTOTPConfiguration;
    handleRefresh: () => void;
    handleDelete: () => void;
}

const TOTPConfiguration = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Fragment>
            <Paper variant="outlined">
                <Box sx={{ p: 3 }}>
                    <Stack direction={"row"} spacing={1} alignItems={"center"}>
                        <QrCode2 fontSize="large" />
                        <Stack spacing={0} sx={{ minWidth: 400 }}>
                            <Box>
                                <Typography display={"inline"} sx={{ fontWeight: "bold" }}>
                                    {props.config.issuer}
                                </Typography>
                                <Typography display={"inline"} variant={"body2"}>
                                    {" (" +
                                        translate("{{algorithm}}, {{digits}} digits, {{seconds}} seconds", {
                                            algorithm: toAlgorithmString(props.config.algorithm),
                                            digits: props.config.digits,
                                            seconds: props.config.period,
                                        }) +
                                        ")"}
                                </Typography>
                            </Box>
                            <Typography variant={"caption"}>
                                {translate("Added when", {
                                    when: props.config.created_at,
                                    formatParams: { when: FormatDateHumanReadable },
                                })}
                            </Typography>
                            <Typography variant={"caption"}>
                                {props.config.last_used_at === undefined
                                    ? translate("Never used")
                                    : translate("Last Used when", {
                                          when: props.config.last_used_at,
                                          formatParams: { when: FormatDateHumanReadable },
                                      })}
                            </Typography>
                        </Stack>
                        <Tooltip title={translate("Remove the Time-based One Time Password configuration")}>
                            <IconButton color="primary" onClick={props.handleDelete}>
                                <DeleteIcon />
                            </IconButton>
                        </Tooltip>
                    </Stack>
                </Box>
            </Paper>
        </Fragment>
    );
};

export default TOTPConfiguration;
