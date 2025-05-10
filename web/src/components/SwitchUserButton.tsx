import React from "react";

import SignOutButton from "@components/SignOutButton";

export interface Props {}

const SwitchUserButton = function (props: Props) {
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
