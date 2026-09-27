// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import SignOutButton from "@components/SignOutButton";

const SwitchUserButton = function () {
    return (
        <SignOutButton
            id={"switch-user-button"}
            text={"Switch User"}
            tooltip={"Logout and continue the current flow"}
            preserve={true}
        />
    );
};

export default SwitchUserButton;
