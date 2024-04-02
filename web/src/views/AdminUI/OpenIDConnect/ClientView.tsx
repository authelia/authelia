import React, { Fragment, useState } from "react";

import { ClientType, OpenIDConnectClient } from "@models/OpenIDConnect";
import ClientItem from "@views/AdminUI/OpenIDConnect/ClientItem";

//import { useTranslation } from "react-i18next";
export interface Props {}

const ClientView = function (props: Props) {
    //const { t: translate } = useTranslation("admin");

    const [clients, setClients] = useState<OpenIDConnectClient[]>([
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
    ]);

    const handleDelete = (index: number) => {
        const updatedClients = [...clients];
        console.log(`delete: ${clients[index].Name}`);
        const filteredClients = updatedClients.filter((_: any, i: any) => i !== index);
        setClients(filteredClients);
    };
    const handleChange = (index: number, updatedClient: OpenIDConnectClient) => {
        const updatedClients = [...clients];
        console.log(`change: client ${updatedClient} at ${index}`);
        updatedClients[index] = updatedClient;
        setClients(updatedClients);
    };

    return (
        <Fragment>
            {clients.map((client, index) => (
                <ClientItem
                    index={index}
                    client={client}
                    description="This is a temporary description!"
                    handleChange={handleChange}
                    handleDelete={handleDelete}
                />
            ))}
        </Fragment>
    );
};

export default ClientView;
