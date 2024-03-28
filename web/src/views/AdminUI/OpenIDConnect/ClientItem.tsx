import React, { Fragment } from "react";
//import React, { Fragment, useState } from "react";

//import { useTranslation } from "react-i18next";

//import { OpenIDConnectClient } from "@models/OpenIDConnect";
import AccordionItem from "@views/AdminUI/AccordionItem";

interface Props {
    index: number;
    name: string;
    description: string;
    //client: OpenIDConnectClient;
    //request_uris: RedirectURIs[];
    secret_last_changed: Date;
    created_at: Date;
    handleInformation: (index: number) => void;
    handleEdit: (index: number) => void;
    handleDelete: (index: number) => void;
}

const ClientItem = function (props: Props) {
    //const { t: translate } = useTranslation("admin");

    //const [showDialogDetails, setShowDialogDetails] = useState<boolean>(false);

    return (
        <Fragment>
            <AccordionItem
                id={`oidc-client-${props.index}`}
                name={props.name}
                description={props.description}
                //name={props.client.name}
            ></AccordionItem>
        </Fragment>
    );
};

export default ClientItem;
