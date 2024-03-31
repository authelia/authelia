import React, { Fragment } from "react";

import { ClientType, OpenIDConnectClient } from "@models/OpenIDConnect";
import ClientItem from "@views/AdminUI/OpenIDConnect/ClientItem";

//import { useTranslation } from "react-i18next";
export interface Props {}

const ClientView = function (props: Props) {
    //const { t: translate } = useTranslation("admin");

    const clients: OpenIDConnectClient[] = [
        {
            ID: "MMh9Xh7R2zUXKBtvCUFpKE9DBsKsO3zAP4HxdPkUybpRApoOzK6UyqHZfysa9eYW7d2x57nfRwDKfm39V5CXZcFSeYK1tpRQhUDt",
            Name: "A Friendly Name for a Client",
            ClientType: ClientType.Confidential,
            RedirectURIs: ["https://example.com/redirect", "https://example.com/redirect2"],
            Audience: ["https://aud1.example.com", "https://aud2.example.com"],
            Scopes: ["scope1", "scope2"],
            AuthorizationPolicy: {
                Name: "Policy1",
                DefaultPolicy: 1,
                Rules: [],
            },
        },
        {
            ID: "MMh9Xh7R2zUXKBtvCUFpKE9DBsKsO3zAP4HxdPkUybpRApoOzK6UyqHZfysa9eYW7d2x57nfRwDKfm39V5CXZcFSeYK1tpRQhUDt",
            Name: "A Friendly Name for a Client",
            ClientType: ClientType.Confidential,
            RedirectURIs: ["https://example.com/redirect", "https://example.com/redirect2"],
            Audience: ["https://aud1.example.com", "https://aud2.example.com"],
            Scopes: ["scope1", "scope2"],
            AuthorizationPolicy: {
                Name: "Policy1",
                DefaultPolicy: 1,
                Rules: [],
            },
        },
        {
            ID: "MMh9Xh7R2zUXKBtvCUFpKE9DBsKsO3zAP4HxdPkUybpRApoOzK6UyqHZfysa9eYW7d2x57nfRwDKfm39V5CXZcFSeYK1tpRQhUDt",
            Name: "A Friendly Name for a Client",
            ClientType: ClientType.Confidential,
            RedirectURIs: ["https://example.com/redirect", "https://example.com/redirect2"],
            Audience: ["https://aud1.example.com", "https://aud2.example.com"],
            Scopes: ["scope1", "scope2"],
            AuthorizationPolicy: {
                Name: "Policy1",
                DefaultPolicy: 1,
                Rules: [],
            },
        },
    ];

    return (
        <Fragment>
            {clients.map((client, index) => (
                <ClientItem
                    index={index}
                    client={client}
                    description="This is a temporary description!"
                    handleInformation={function (index: number): void {
                        throw new Error("Function not implemented.");
                    }}
                    handleEdit={function (index: number): void {
                        throw new Error("Function not implemented.");
                    }}
                    handleDelete={function (index: number): void {
                        throw new Error("Function not implemented.");
                    }}
                />
            ))}
        </Fragment>
    );
};

export default ClientView;
