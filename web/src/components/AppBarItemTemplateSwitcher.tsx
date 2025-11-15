import React, { Fragment, useMemo, useState } from "react";

import { AutoAwesome } from "@mui/icons-material";
import { Box, IconButton, ListItemText, Menu, MenuItem, Tooltip, Typography, useTheme } from "@mui/material";

import { usePortalTemplate } from "@contexts/PortalTemplateContext";
import { PortalTemplateName } from "@themes/portalTemplates";

const AppBarItemTemplateSwitcher = () => {
    const theme = useTheme();
    const { allowSwitcher, availableTemplates, template, switchTemplate } = usePortalTemplate();
    const [anchorEl, setAnchorEl] = useState<HTMLElement | null>(null);
    const open = Boolean(anchorEl);

    const options = useMemo(() => availableTemplates, [availableTemplates]);

    if (!allowSwitcher || options.length <= 1) {
        return null;
    }

    const handleClick = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget);
    };

    const handleClose = () => {
        setAnchorEl(null);
    };

    const handleSelect = (name: PortalTemplateName) => {
        switchTemplate(name);
        handleClose();
    };

    const activeTemplate = options.find((option) => option.name === template);

    return (
        <Fragment>
            <Box sx={{ display: "flex", alignItems: "center", textAlign: "center" }}>
                <Tooltip title="Switch portal theme">
                    <IconButton
                        id="template-switcher-button"
                        onClick={handleClick}
                        sx={{ ml: 2 }}
                        aria-controls={open ? "template-switcher-menu" : undefined}
                        aria-expanded={open ? "true" : undefined}
                        aria-haspopup="true"
                    >
                        <AutoAwesome />
                        <Typography sx={{ paddingLeft: theme.spacing(1) }}>
                            {activeTemplate?.displayName ?? "Templates"}
                        </Typography>
                    </IconButton>
                </Tooltip>
            </Box>
            <Menu
                anchorEl={anchorEl}
                id="template-switcher-menu"
                open={open}
                onClose={handleClose}
                slotProps={{
                    list: {
                        "aria-labelledby": "template-switcher-button",
                    },
                    paper: {
                        elevation: 0,
                        sx: {
                            maxHeight: { xs: "80vh", sm: "70vh", md: "50vh", lg: "40vh" },
                            filter: "drop-shadow(0px 2px 8px rgba(0,0,0,0.32))",
                            "&::before": {
                                content: '""',
                                position: "relative",
                                top: 0,
                                right: 14,
                                width: 10,
                                height: 10,
                                bgcolor: "background.paper",
                                transform: "translateY(-50%) rotate(45deg)",
                                zIndex: 0,
                            },
                        },
                    },
                }}
            >
                {options.map((option) => (
                    <MenuItem
                        key={option.name}
                        id={`portal-template-${option.name}`}
                        selected={option.name === template}
                        onClick={() => handleSelect(option.name)}
                    >
                        <ListItemText primary={option.displayName} secondary={option.description} />
                    </MenuItem>
                ))}
            </Menu>
        </Fragment>
    );
};

export default AppBarItemTemplateSwitcher;
