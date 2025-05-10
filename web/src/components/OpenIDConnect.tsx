import {
    AccountBox,
    Autorenew,
    Contacts,
    Drafts,
    Group,
    Home,
    LockOpen,
    PhoneAndroid,
    Policy,
} from "@mui/icons-material";

import {
    ScopeAddress,
    ScopeAutheliaBearerAuthz,
    ScopeEmail,
    ScopeGroups,
    ScopeOfflineAccess,
    ScopeOpenID,
    ScopePhone,
    ScopeProfile,
} from "@constants/OpenIDConnect";

export function ScopeAvatar(scope: string) {
    switch (scope) {
        case ScopeOpenID:
            return <AccountBox />;
        case ScopeOfflineAccess:
            return <Autorenew />;
        case ScopeProfile:
            return <Contacts />;
        case ScopeGroups:
            return <Group />;
        case ScopeEmail:
            return <Drafts />;
        case ScopePhone:
            return <PhoneAndroid />;
        case ScopeAddress:
            return <Home />;
        case ScopeAutheliaBearerAuthz:
            return <LockOpen />;
        default:
            return <Policy />;
    }
}

export function ScopeDescription(scope: string): string {
    switch (scope) {
        case ScopeOpenID:
            return "Use OpenID to verify your identity";
        case ScopeOfflineAccess:
            return "Automatically refresh these permissions without user interaction";
        case ScopeProfile:
            return "Access your profile information";
        case ScopeGroups:
            return "Access your group membership";
        case ScopeEmail:
            return "Access your email addresses";
        case ScopePhone:
            return "Access your phone number";
        case ScopeAddress:
            return "Access your address";
        case ScopeAutheliaBearerAuthz:
            return "Access protected resources logged in as you";
        default:
            return scope;
    }
}
