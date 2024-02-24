import React, { useState } from "react";

import ExpandLess from "@mui/icons-material/ExpandLess";
import ExpandMore from "@mui/icons-material/ExpandMore";
import LanguageIcon from "@mui/icons-material/Language";
import { Box, Collapse, IconButton, ListItemText, ListSubheader, Menu, MenuItem, Select, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";

import { supportedLngsNames } from "i18n";

export interface Props {
    value: string;
    onChange: Function;
    picker: Boolean;
}

const languageTree = supportedLngsNames
    .filter((lng: any) => !lng.parent)
    .map((lng) => {
        return {
            name: lng.name,
            lng: lng.lng,
            children: supportedLngsNames.filter((l: any) => l.parent === lng.lng),
        };
    });

const LanguageSelector = function (props: Props) {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl);
    const [expanded, setExpanded] = useState("");

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

    const handleChange = (lng: string) => {
        setAnchorEl(null);
        if (lng) {
            props.onChange(lng);
        }
    };

    const handleCollapse = (locale: string) => {
        if (locale === expanded) {
            setExpanded("");
        } else {
            setExpanded(locale);
        }
    };

    const languages = languageTree.map((lng) => {
        // if locale have not children, it is selectable
        if (lng.children.length === 0) {
            return (
                <MenuItem key={lng.lng} onClick={() => handleChange(lng.lng)} value={lng.lng}>
                    <ListItemText>{lng.name}</ListItemText>
                </MenuItem>
            );
        } else if (lng.children.length === 1) {
            // if the locale have only one child, we select the children
            return (
                <MenuItem key={lng.lng} onClick={() => handleChange(lng.lng)} value={lng.lng}>
                    <ListItemText>{lng.name}</ListItemText>
                </MenuItem>
            );
        } else {
            // if the locale have more than 1 children they are added
            const children = lng.children.map((child) => {
                return (
                    <MenuItem key={child.lng} onClick={() => handleChange(child.lng)} value={child.lng}>
                        <ListItemText>&nbsp;&nbsp;{child.name}</ListItemText>
                    </MenuItem>
                );
            });

            if (props.picker) {
                return (
                    <div key={lng.lng}>
                        <MenuItem value={lng.lng}>
                            <ListItemText onClick={() => handleCollapse(lng.lng)}>{lng.name}</ListItemText>
                            {expanded === lng.lng ? <ExpandLess /> : <ExpandMore />}
                        </MenuItem>
                        <Collapse in={expanded === lng.lng} timeout="auto">
                            {children}
                        </Collapse>
                    </div>
                );
            } else {
                children.unshift(<ListSubheader>{lng.name}</ListSubheader>);
                return children;
            }
        }
    });

    return props.picker ? (
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
                onClose={() => handleChange("")}
                transformOrigin={{ horizontal: "right", vertical: "top" }}
                anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
                PaperProps={{
                    style: {
                        width: 350,
                    },
                }}
            >
                {languages}
            </Menu>
        </Box>
    ) : (
        <Select value={props.value}>{languages}</Select>
    );
};

LanguageSelector.defaultProps = {};

export default LanguageSelector;
