import { Fragment } from "react";

import {
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    Divider,
    Typography,
} from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";

import { FormatDateHumanReadable } from "@i18n/formats";
import { UserInfoTOTPConfiguration, toAlgorithmString } from "@models/TOTPConfiguration";

interface Props {
    config: null | undefined | UserInfoTOTPConfiguration;
    open: boolean;
    handleClose: () => void;
}

const OneTimePasswordInformationDialog = function (props: Props) {
    const { t: translate } = useTranslation("settings");

    return (
        <Dialog open={props.open} onClose={props.handleClose} aria-labelledby="one-time-password-info-dialog-title">
            <DialogTitle id="one-time-password-info-dialog-title">
                {translate("One-Time Password Information")}
            </DialogTitle>
            <DialogContent>
                {props.config ? (
                    <Fragment>
                        <DialogContentText sx={{ mb: 3 }}>
                            {translate("Extended information for One-Time Password")}
                        </DialogContentText>
                        <Grid container spacing={2}>
                            <Grid size={{ md: 3 }} sx={{ display: { md: "block", xs: "none" } }} />
                            <Grid size={{ xs: 12 }}>
                                <Divider />
                            </Grid>
                            <PropertyText
                                name={translate("Algorithm")}
                                value={translate("{{algorithm}}", {
                                    algorithm: toAlgorithmString(props.config.algorithm),
                                })}
                            />
                            <PropertyText
                                name={translate("Digits")}
                                value={translate("{{digits}}", {
                                    digits: props.config.digits,
                                })}
                            />
                            <PropertyText
                                name={translate("Period")}
                                value={translate("{{seconds}}", {
                                    seconds: props.config.period,
                                })}
                            />
                            <PropertyText name={translate("Issuer")} value={props.config.issuer} />
                            <PropertyText
                                name={translate("Added")}
                                value={translate("{{when, datetime}}", {
                                    formatParams: { when: FormatDateHumanReadable },
                                    when: new Date(props.config.created_at),
                                })}
                            />
                            <PropertyText
                                name={translate("Last Used")}
                                value={
                                    props.config.last_used_at
                                        ? translate("{{when, datetime}}", {
                                              formatParams: { when: FormatDateHumanReadable },
                                              when: new Date(props.config.last_used_at),
                                          })
                                        : translate("Never")
                                }
                            />
                        </Grid>
                    </Fragment>
                ) : (
                    <DialogContentText sx={{ mb: 3 }}>
                        {translate("The One-Time Password information is not loaded")}
                    </DialogContentText>
                )}
            </DialogContent>
            <DialogActions>
                <Button id={"dialog-close"} onClick={props.handleClose}>
                    {translate("Close")}
                </Button>
            </DialogActions>
        </Dialog>
    );
};

interface PropertyTextProps {
    readonly name: string;
    readonly value: string;
    readonly xs?: number;
}

function PropertyText(props: PropertyTextProps) {
    return (
        <Grid size={{ xs: props.xs ?? 12 }}>
            <Typography display="inline" sx={{ fontWeight: "bold" }}>
                {`${props.name}: `}
            </Typography>
            <Typography display="inline">{props.value}</Typography>
        </Grid>
    );
}

export default OneTimePasswordInformationDialog;
