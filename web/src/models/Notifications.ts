export type NotificationLevel = "error" | "info" | "success" | "warning";

export interface Notification {
    message: string;
    level: NotificationLevel;
    timeout: number;
}
