import React, { Fragment, useEffect, useState } from "react";

import { useTranslation } from "react-i18next";

import { useNotifications } from "@hooks/NotificationsContext";
import { useOpenIDConnectClients } from "@hooks/OIDCClientConfig";
import { ClientAuthorizationPolicyRuleImpl, ExistingScopes, OpenIDConnectClient } from "@models/OpenIDConnect";
import ClientItem from "@views/AdminUI/OpenIDConnect/ClientItem";

export interface Props {}

const ClientView = function (props: Props) {
    const { t: translate } = useTranslation("admin");

    const [clients, setClients] = useState<OpenIDConnectClient[]>([
        {
            ID: "MMh9Xh7R2zUXKBtvCUFpKE9DBsKsO3zAP4HxdPkUybpRApoOzK6UyqHZfysa9eYW7d2x57nfRwDKfm39V5CXZcFSeYK1tpRQhUDt",
            Name: "This is one client",
            Public: false,
            RedirectURIs: ["https://example.com/redirect", "https://example.com/redirect2"],
            Audience: ["https://aud1.example.com", "https://aud2.example.com"],
            Scopes: [ExistingScopes.openid, ExistingScopes.email],
            DefaultAuthorizationPolicy: "one_factor",
            AuthorizationPolicy: {
                Name: "Policy1",

                Rules: [
                    (() => {
                        return new ClientAuthorizationPolicyRuleImpl("user:john");
                    })(),
                ],
            },
        },
        {
            ID: "MMh9Xh7R2zUXKBtvCUFpKE9DBsKsO3zAP4HxdPkUybpRApoOzK6UyqHZfysa9eYW7d2x57nfRwDKfm39V5CXZcFSeYK1tpRQhUDt",
            Name: "Another client",
            Public: true,
            RedirectURIs: ["https://example.com/redirect", "https://example.com/redirect2"],
            Audience: ["https://aud1.example.com", "https://aud2.example.com"],
            Scopes: [ExistingScopes.offline_access, ExistingScopes.email],
            DefaultAuthorizationPolicy: "one_factor",
            AuthorizationPolicy: {
                Name: "Policy2",
                Rules: [
                    (() => {
                        return new ClientAuthorizationPolicyRuleImpl("group:services");
                    })(),
                ],
            },
        },
        {
            ID: "MMh9Xh7R2zUXKBtvCUFpKE9DBsKsO3zAP4HxdPkUybpRApoOzK6UyqHZfysa9eYW7d2x57nfRwDKfm39V5CXZcFSeYK1tpRQhUDt",
            Name: "A third client",
            Public: false,
            RedirectURIs: ["https://example.com/redirect", "https://example.com/redirect2"],
            Audience: ["https://aud1.example.com", "https://aud2.example.com"],
            Scopes: [ExistingScopes.profile, ExistingScopes.email],
            DefaultAuthorizationPolicy: "one_factor",
            AuthorizationPolicy: {
                Name: "Policy3",
                Rules: [
                    (() => {
                        return new ClientAuthorizationPolicyRuleImpl("user:steve");
                    })(),
                ],
            },
        },
    ]);

    const { createErrorNotification } = useNotifications();
    const [openIDConnectClients, fetchOpenIDConnectClients, , fetchOpenIDConnectClientsError] =
        useOpenIDConnectClients();

    useEffect(() => {
        if (fetchOpenIDConnectClientsError) {
            createErrorNotification(
                translate("There was an issue retrieving the {{item}}", {
                    item: translate("OpenIDConnect Clients"),
                }),
            );
        }
    }, [fetchOpenIDConnectClientsError, createErrorNotification, translate]);

    useEffect(() => {
        if (openIDConnectClients === undefined) {
            return;
        }
        setClients(openIDConnectClients);
        console.log(openIDConnectClients);
    }, [openIDConnectClients]);

    useEffect(() => {
        fetchOpenIDConnectClients();
    }, [fetchOpenIDConnectClients]);

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
                    key={index}
                    index={index}
                    client={client}
                    handleChange={handleChange}
                    handleDelete={handleDelete}
                />
            ))}
        </Fragment>
    );
};

export default ClientView;
