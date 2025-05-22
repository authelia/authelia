import React from "react";

import SignOutButton from "@components/SignOutButton";

export interface Props {}

const LogoutButton = function (props: Props) {
    return <SignOutButton id={"logout-button"} text={"Logout"} tooltip={"Logout and clear any current flow"} />;
};

export default LogoutButton;
