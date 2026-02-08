import { FC, Fragment, useEffect, useState } from "react";

import { Checkbox, FormControlLabel, Tooltip } from "@mui/material";
import Grid from "@mui/material/Grid";
import { useTranslation } from "react-i18next";
import { makeStyles } from "tss-react/mui";

export interface Props {
    pre_configuration: boolean;
    onChangePreConfiguration: (_value: boolean) => void;
}

const DecisionFormPreConfiguration: FC<Props> = (props: Props) => {
    const { t: translate } = useTranslation(["consent"]);

    const { classes } = useStyles();

    const [preConfigure, setPreConfigure] = useState(false);

    const handlePreConfigureChanged = () => {
        setPreConfigure((preConfigure) => !preConfigure);
    };

    useEffect(() => {
        props.onChangePreConfiguration(preConfigure);
    }, [preConfigure, props]);

    return (
        <Fragment>
            {props.pre_configuration ? (
                <Grid size={{ xs: 12 }}>
                    <Tooltip title={translate("This saves this consent as a pre-configured consent for future use")}>
                        <FormControlLabel
                            control={
                                <Checkbox
                                    id="pre-configure"
                                    checked={preConfigure}
                                    onChange={handlePreConfigureChanged}
                                    value="preConfigure"
                                    color="primary"
                                />
                            }
                            className={classes.preConfigure}
                            label={translate("Remember Consent")}
                        />
                    </Tooltip>
                </Grid>
            ) : null}
        </Fragment>
    );
};

const useStyles = makeStyles()(() => ({
    preConfigure: {},
}));

export default DecisionFormPreConfiguration;
