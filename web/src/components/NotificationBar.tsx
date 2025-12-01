import { useEffect, useState } from "react";

import { Box } from "@mui/material";
import { FiAlertCircle, FiCheckCircle, FiInfo } from "react-icons/fi";

import { useNotifications } from "@hooks/NotificationsContext";

export interface Props {
    onClose: () => void;
}

const NotificationBar = function (props: Props) {
    const { notification } = useNotifications();
    const [visible, setVisible] = useState(false);

    useEffect(() => {
        if (notification) {
            setVisible(true);
            const timer = setTimeout(() => {
                setVisible(false);
                setTimeout(() => props.onClose(), 250);
            }, notification.timeout * 1000 || 3000);

            return () => clearTimeout(timer);
        }
    }, [notification, props]);

    if (!notification) return null;

    const getIcon = () => {
        switch (notification.level) {
            case "success":
                return <FiCheckCircle className="w-5 h-5" />;
            case "error":
                return <FiAlertCircle className="w-5 h-5" />;
            case "info":
            default:
                return <FiInfo className="w-5 h-5" />;
        }
    };

    const getColors = () => {
        switch (notification.level) {
            case "success":
                return {
                    border: "rgba(48, 164, 108, 0.4)",
                    iconColor: "#30a46c",
                };
            case "error":
                return {
                    border: "rgba(220, 53, 69, 0.4)",
                    iconColor: "#ff6b6b",
                };
            case "info":
            default:
                return {
                    border: "rgba(42, 162, 193, 0.4)",
                    iconColor: "#2aa2c1",
                };
        }
    };

    const colors = getColors();

    return (
        <Box
            sx={{
                position: "fixed",
                top: "8rem",
                right: "2.5rem",
                zIndex: 9999,
                pointerEvents: "none",
            }}
        >
            <Box
                sx={{
                    pointerEvents: "auto",
                    maxWidth: "20rem",
                    width: "auto",
                    background: "#1e2b39",
                    border: `1px solid ${colors.border}`,
                    borderRadius: "0.5rem",
                    padding: "1rem",
                    display: "flex",
                    alignItems: "flex-start",
                    gap: "0.75rem",
                    boxShadow: "0 4px 6px -1px rgba(0, 0, 0, 0.1), 0 2px 4px -1px rgba(0, 0, 0, 0.06)",
                    opacity: visible ? 1 : 0,
                    transform: visible ? "translateY(0)" : "translateY(-4px)",
                    transition: "opacity 200ms ease-in-out, transform 200ms ease-in-out",
                }}
            >
                <Box
                    sx={{
                        color: colors.iconColor,
                        flexShrink: 0,
                        marginTop: "0.125rem",
                    }}
                >
                    {getIcon()}
                </Box>
                <Box
                    sx={{
                        flex: 1,
                        color: "hsla(0, 0%, 100%, 0.74)",
                        fontSize: "0.875rem",
                        lineHeight: 1.5,
                        wordBreak: "break-word",
                    }}
                >
                    {notification.message}
                </Box>
            </Box>
        </Box>
    );
};

export default NotificationBar;

<style>
{`
.w-5 {
  width: 1.25rem;
  height: 1.25rem;
}

.h-5 {
  height: 1.25rem;
}
`}
</style>
