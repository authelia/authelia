// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { useState } from "react";

import { IdCardLanyard } from "lucide-react";

import DiscordSvg from "@assets/images/identity/discord.svg?react";
import GitHubSvg from "@assets/images/identity/github.svg?react";
import OpenIDSvg from "@assets/images/identity/openid.svg?react";

export interface Props {
    type: string;

    className?: string;
    logoURI?: string;
}

const marks: { [type: string]: typeof DiscordSvg } = {
    discord: DiscordSvg,
    github: GitHubSvg,
    openid_connect: OpenIDSvg,
};

// ExternalIdentityIcon renders the icon of an external identity provider. The logo the provider is configured with is
// preferred, falling back to the mark bundled for its type, and finally to a generic icon for a type this build has no
// mark for, which a newer backend may serve to an older login page. A logo which fails to load
// falls back the same way rather than leaving a broken image on the sign in button. The URI which failed is recorded
// rather than the fact of the failure, so a logo which changes is attempted again.
const ExternalIdentityIcon = function (props: Props) {
    const [failed, setFailed] = useState<string | undefined>(undefined);

    const className = props.className ?? "mr-2 h-5 w-5 shrink-0";

    if (props.logoURI && props.logoURI !== failed) {
        return (
            <img
                src={props.logoURI}
                alt=""
                aria-hidden
                className={className}
                onError={() => setFailed(props.logoURI)}
            />
        );
    }

    const Mark = marks[props.type];

    if (Mark) {
        return <Mark aria-hidden className={className} />;
    }

    return <IdCardLanyard aria-hidden className={className} />;
};

export default ExternalIdentityIcon;
