import React, { useState } from "react";

import { Brightness4, Brightness5, Brightness6, BrightnessAuto, BrightnessHigh } from "@mui/icons-material";
import { IconButton, Menu, MenuItem, useTheme } from "@mui/material";

import { useThemeContext } from "@contexts/ThemeContext";
import { ThemeNameAuto, ThemeNameDark, ThemeNameGrey, ThemeNameLight, ThemeNameOled } from "@themes/index";

const ThemeChanger: React.FC = () => {
    const { themeName, setThemeName } = useThemeContext();
    const theme = useTheme();
    const [anchorElement, setAnchorElement] = useState<null | HTMLElement>(null);

    const themes = [
        { name: "Auto", value: ThemeNameAuto, icon: BrightnessAuto },
        { name: "Light", value: ThemeNameLight, icon: BrightnessHigh },
        { name: "Grey", value: ThemeNameGrey, icon: Brightness6 },
        { name: "Dark", value: ThemeNameDark, icon: Brightness4 },
        { name: "Oled", value: ThemeNameOled, icon: Brightness5 },
    ];

    const CurrentIcon = themes.find((t) => t.value === themeName)?.icon || BrightnessAuto;

    const handleClick = (event: React.MouseEvent<HTMLButtonElement>) => {
        setAnchorElement(event.currentTarget);
    };

    const handleClose = () => {
        setAnchorElement(null);
    };

    const handleThemeChange = (newTheme: string) => {
        setThemeName(newTheme);
        handleClose();
    };

    return (
        <>
            <IconButton
                onClick={handleClick}
                aria-label="change theme"
                aria-controls="theme-menu"
                aria-haspopup="true"
                sx={{
                    color: theme.palette.text.primary,
                }}
            >
                <CurrentIcon />
            </IconButton>
            <Menu
                id="theme-menu"
                anchorEl={anchorElement}
                keepMounted
                open={Boolean(anchorElement)}
                onClose={handleClose}
            >
                {themes.map((t) => (
                    <MenuItem key={t.value} onClick={() => handleThemeChange(t.value)} selected={themeName === t.value}>
                        <t.icon sx={{ mr: 1 }} />
                        {t.name}
                    </MenuItem>
                ))}
            </Menu>
        </>
    );
};

export default ThemeChanger;
