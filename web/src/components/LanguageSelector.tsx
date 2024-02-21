import React, { useState } from "react";

import LanguageIcon from "@mui/icons-material/Language";
import { Box, IconButton, Menu, MenuItem, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";

import { supportedLngsNames } from "i18n";

export interface Props {
    value: string;
    onChange: Function;
}

const LanguageSelector = function (props: Props) {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl);

    const styles = makeStyles((theme: Theme) => ({
        topRight: {
            position: "absolute",
            top: "20px",
            right: "20px",
        },
    }))();

    const handleClick = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget);
    };
    const handleClose = (lng: string) => {
        setAnchorEl(null);
        if (lng !== "") {
            props.onChange(lng);
        }
    };

    const languages = supportedLngsNames.map((lng) => (
        <MenuItem key={lng.lng} onClick={() => handleClose(lng.lng)} value={lng.lng}>
            {lng.name}
        </MenuItem>
    ));

    return (
        <Box className={classnames(styles.topRight)}>
            <IconButton
                onClick={handleClick}
                size="small"
                sx={{ ml: 2 }}
                aria-controls={open ? "account-menu" : undefined}
                aria-haspopup="true"
                aria-expanded={open ? "true" : undefined}
            >
                <LanguageIcon></LanguageIcon> <span>{props.value}</span>
            </IconButton>
            <Menu
                anchorEl={anchorEl}
                id="account-menu"
                open={open}
                onClose={() => handleClose("")}
                onClick={() => handleClose("")}
                transformOrigin={{ horizontal: "right", vertical: "top" }}
                anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
            >
                {languages}
            </Menu>
        </Box>
    );
};

LanguageSelector.defaultProps = {};

export default LanguageSelector;
