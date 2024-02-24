import React, { useCallback, useEffect, useState } from "react";

import ExpandLess from "@mui/icons-material/ExpandLess";
import ExpandMore from "@mui/icons-material/ExpandMore";
import LanguageIcon from "@mui/icons-material/Language";
import { Box, Collapse, IconButton, ListItemText, ListSubheader, Menu, MenuItem, Select, Theme } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";
import classnames from "classnames";

import { Language } from "@models/LocaleInformation";

export interface Props {
    value: string;
    localeList: Array<Language>;
    onChange: Function;
    picker: Boolean;
}

const LanguageSelector = function (props: Props) {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const open = Boolean(anchorEl);
    const [expanded, setExpanded] = useState("");
    const [menuItems, setMenuItems] = useState<null | Array<any>>(null);

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

    const { onChange, picker: pickerMode } = props;

    const closeMenu = useCallback(() => {
        setAnchorEl(null);
    }, []);

    const handleChange = useCallback(
        (lng: string) => {
            closeMenu();
            if (lng) {
                onChange(lng);
            }
        },
        [onChange, closeMenu],
    );

    const handleCollapse = useCallback(
        (locale: string) => {
            if (locale === expanded) {
                setExpanded("");
            } else {
                setExpanded(locale);
            }
        },
        [expanded],
    );

    // convert the language list into a tree with base languages and their childs
    const localeInfoToTree = useCallback(() => {
        const tree = props.localeList
            .filter((lng: any) => !lng.parent)
            .map((lng) => {
                return {
                    display: lng.display,
                    locale: lng.locale,
                    children: props.localeList.filter((l: any) => l.parent === lng.locale),
                };
            });
        return tree;
    }, [props.localeList]);

    const generateMenuItems = useCallback(() => {
        const tree = localeInfoToTree();
        const items = tree.map((lng) => {
            // if locale have not children, it is selectable
            if (lng.children.length === 0) {
                return (
                    <MenuItem key={lng.locale} onClick={() => handleChange(lng.locale)} value={lng.locale}>
                        <ListItemText>{lng.display}</ListItemText>
                    </MenuItem>
                );
            } else if (lng.children.length === 1) {
                // if the locale have only one child, we select the children
                return (
                    <MenuItem key={lng.locale} onClick={() => handleChange(lng.locale)} value={lng.locale}>
                        <ListItemText>{lng.display}</ListItemText>
                    </MenuItem>
                );
            } else {
                // if the locale have more than 1 children they are added
                const children = lng.children.map((child) => {
                    return (
                        <MenuItem key={child.locale} onClick={() => handleChange(child.locale)} value={child.locale}>
                            <ListItemText>&nbsp;&nbsp;{child.display}</ListItemText>
                        </MenuItem>
                    );
                });

                if (pickerMode) {
                    return (
                        <div key={lng.locale}>
                            <MenuItem value={lng.locale}>
                                <ListItemText onClick={() => handleCollapse(lng.locale)}>{lng.display}</ListItemText>
                                {expanded === lng.locale ? <ExpandLess /> : <ExpandMore />}
                            </MenuItem>
                            <Collapse in={expanded === lng.locale} timeout="auto">
                                {children}
                            </Collapse>
                        </div>
                    );
                } else {
                    children.unshift(<ListSubheader>{lng.display}</ListSubheader>);
                    return children;
                }
            }
        });
        setMenuItems(items);
    }, [localeInfoToTree, expanded, pickerMode, handleCollapse, handleChange]);

    useEffect(() => {
        generateMenuItems();
    }, [generateMenuItems]);

    return pickerMode ? (
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
                onClose={closeMenu}
                transformOrigin={{ horizontal: "right", vertical: "top" }}
                anchorOrigin={{ horizontal: "right", vertical: "bottom" }}
                PaperProps={{
                    style: {
                        width: 350,
                    },
                }}
            >
                {menuItems}
            </Menu>
        </Box>
    ) : (
        <Select value={props.value}>{/*languages*/}</Select>
    );
};

LanguageSelector.defaultProps = {};

export default LanguageSelector;
