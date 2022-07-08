import { AlertColor } from "@mui/material";

export interface Notification {
    message: string;
    level: AlertColor;
    timeout: number;
}
