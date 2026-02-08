import SignOutButton from "@components/SignOutButton";

const LogoutButton = function () {
    return <SignOutButton id={"logout-button"} text={"Logout"} tooltip={"Logout and clear any current flow"} />;
};

export default LogoutButton;
