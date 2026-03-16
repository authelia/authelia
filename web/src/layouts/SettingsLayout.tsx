import { ReactNode, SyntheticEvent, useCallback, useEffect, useState } from "react";

import { LayoutDashboard, Menu, Shield, ShieldCheck, X } from "lucide-react";
import { useTranslation } from "react-i18next";

import { Button } from "@components/UI/Button";
import { Separator } from "@components/UI/Separator";
import { Sheet, SheetContent } from "@components/UI/Sheet";
import { EncodedName } from "@constants/constants";
import {
    IndexRoute,
    SecuritySubRoute,
    SettingsRoute,
    SettingsTwoFactorAuthenticationSubRoute,
} from "@constants/Routes";
import { useRouterNavigate } from "@hooks/RouterNavigate";
import { cn } from "@utils/Styles";

export interface Props {
    children?: ReactNode;
    drawerWidth?: number;
}

const defaultDrawerWidth = 240;

const SettingsLayout = function (props: Props) {
    const { t: translate } = useTranslation("settings");
    const [drawerOpen, setDrawerOpen] = useState(false);

    useEffect(() => {
        document.title = translate("Settings - {{authelia}}", { authelia: atob(String.fromCodePoint(...EncodedName)) });
    }, [translate]);

    const drawerWidth = props.drawerWidth ?? defaultDrawerWidth;

    const handleToggleDrawer = (event: SyntheticEvent) => {
        if (
            event.nativeEvent instanceof KeyboardEvent &&
            event.nativeEvent.type === "keydown" &&
            (event.nativeEvent.key === "Tab" || event.nativeEvent.key === "Shift")
        ) {
            return;
        }

        setDrawerOpen((state) => !state);
    };

    return (
        <div className="flex">
            <nav className="fixed top-0 z-50 w-full bg-primary text-primary-foreground">
                <div className="flex items-center px-4 py-3">
                    <Button
                        id="settings-menu"
                        variant="ghost"
                        size="icon"
                        aria-label="open drawer"
                        onClick={handleToggleDrawer}
                        className="mr-4 text-primary-foreground hover:bg-primary-foreground/10"
                    >
                        <Menu className="size-6" />
                    </Button>
                    <h6 className={cn("grow text-lg font-medium", drawerOpen && "hidden")}>{translate("Settings")}</h6>
                </div>
            </nav>
            <Sheet open={drawerOpen} onOpenChange={setDrawerOpen}>
                <SheetContent side="left" className="p-0" style={{ width: drawerWidth }}>
                    <div className="text-center" onClick={handleToggleDrawer}>
                        <h6 className="my-4 text-lg font-medium">{translate("Settings")}</h6>
                        <Separator />
                        <ul className="list-none p-0">
                            {navItems.map((item) => (
                                <DrawerNavItem
                                    key={item.keyname}
                                    keyname={item.keyname}
                                    text={translate(item.text)}
                                    pathname={item.pathname}
                                    icon={item.icon}
                                />
                            ))}
                        </ul>
                    </div>
                </SheetContent>
            </Sheet>
            <main className="grow p-0 pt-14 sm:p-6 sm:pt-20">{props.children}</main>
        </div>
    );
};

interface NavItem {
    keyname: string;
    text: string;
    pathname: string;
    icon?: ReactNode;
}

const navItems: NavItem[] = [
    {
        icon: <LayoutDashboard className="size-5 text-primary" />,
        keyname: "overview",
        pathname: SettingsRoute,
        text: "Overview",
    },
    {
        icon: <Shield className="size-5 text-primary" />,
        keyname: "security",
        pathname: `${SettingsRoute}${SecuritySubRoute}`,
        text: "Security",
    },
    {
        icon: <ShieldCheck className="size-5 text-primary" />,
        keyname: "twofactor",
        pathname: `${SettingsRoute}${SettingsTwoFactorAuthenticationSubRoute}`,
        text: "Two-Factor Authentication",
    },
    { icon: <X className="size-5 text-destructive" />, keyname: "close", pathname: IndexRoute, text: "Close" },
];

const DrawerNavItem = function (props: NavItem) {
    const selected =
        globalThis.location.pathname === props.pathname || globalThis.location.pathname === props.pathname + "/";
    const navigate = useRouterNavigate();

    const handleOnClick = useCallback(() => {
        if (selected) {
            return;
        }

        navigate(props.pathname);
    }, [navigate, props, selected]);

    return (
        <li>
            <button
                id={`settings-menu-${props.keyname}`}
                onClick={handleOnClick}
                className={cn(
                    "flex w-full items-center gap-3 px-4 py-3 text-left transition-colors hover:bg-accent",
                    selected && "bg-accent font-medium",
                )}
            >
                {props.icon ? <span className="shrink-0">{props.icon}</span> : null}
                <span>{props.text}</span>
            </button>
        </li>
    );
};

export default SettingsLayout;
