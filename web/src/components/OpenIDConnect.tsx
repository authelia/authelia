import { CircleUserRound, Home, Lock, Mail, Phone, RefreshCw, Shield, UserRound, Users } from "lucide-react";

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
            return <CircleUserRound className="size-5" />;
        case ScopeOfflineAccess:
            return <RefreshCw className="size-5" />;
        case ScopeProfile:
            return <UserRound className="size-5" />;
        case ScopeGroups:
            return <Users className="size-5" />;
        case ScopeEmail:
            return <Mail className="size-5" />;
        case ScopePhone:
            return <Phone className="size-5" />;
        case ScopeAddress:
            return <Home className="size-5" />;
        case ScopeAutheliaBearerAuthz:
            return <Lock className="size-5" />;
        default:
            return <Shield className="size-5" />;
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
