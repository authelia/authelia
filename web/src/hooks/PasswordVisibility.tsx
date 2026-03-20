import { useState } from "react";
import { IconButton, InputAdornment } from "@mui/material";
import { Visibility, VisibilityOff } from "@mui/icons-material";

export const usePasswordVisibility = () => {
    const [showPassword, setShowPassword] = useState(false);

    const getPasswordSlotProps = () => ({
        input: {
            endAdornment: (
                <InputAdornment position="end">
                    <IconButton
                        aria-label="toggle password visibility"
                        edge="end"
                        size="large"
                        onMouseDown={() => setShowPassword(true)}
                        onMouseUp={() => setShowPassword(false)}
                        onMouseLeave={() => setShowPassword(false)}
                        onTouchStart={() => setShowPassword(true)}
                        onTouchEnd={() => setShowPassword(false)}
                        onTouchCancel={() => setShowPassword(false)}
                        onKeyDown={(e) => {
                            if (e.key === " ") {
                                setShowPassword(true);
                                e.preventDefault();
                            }
                        }}
                        onKeyUp={(e) => {
                            if (e.key === " ") {
                                setShowPassword(false);
                                e.preventDefault();
                            }
                        }}
                    >
                        {showPassword ? <Visibility /> : <VisibilityOff />}
                    </IconButton>
                </InputAdornment>
            ),
        },
    });

    return {
        showPassword,
        setShowPassword,
        passwordSlotProps: getPasswordSlotProps(),
    };
};
