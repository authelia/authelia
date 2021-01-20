import { Level } from "../components/ColoredSnackbarContent";

export interface Notification {
    message: string;
    level: Level;
    timeout: number;
}
