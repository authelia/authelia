import React from "react"
import { Redirect } from "react-router";

async function logout() {
    return fetch("/api/logout", {method: "POST"})
}

export default class LogoutView extends React.Component {
    componentDidMount() {
        logout().catch(console.error);
    }

    render() {
        return <Redirect to='/' />;
    }
}