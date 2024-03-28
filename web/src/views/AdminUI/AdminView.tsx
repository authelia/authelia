import React, { Fragment } from "react";

import AccordionItem from "@views/AdminUI/AccordionItem";

//import { useTranslation } from "react-i18next";
export interface Props {}

const AdminView = function (props: Props) {
    //const { t: translate } = useTranslation("admin");

    return (
        <Fragment>
            {/* <AccordionItem
                id={`oidc-client${props.index}`}
                name={props.name}
                handleInformation={handleInformation}
                handleEdit={handleEdit}
                handleDelete={handleDelete}
            ></AccordionItem> */}
            <AccordionItem
                id={`oidc-client1`}
                name="Client1"
                description="This is a temporary description!"
            ></AccordionItem>
            <AccordionItem
                id={`oidc-client2`}
                name="Client2"
                description="This is a temporary description!"
            ></AccordionItem>
            <AccordionItem
                id={`oidc-client3`}
                name="Client3"
                description="This is a temporary description!"
            ></AccordionItem>
            <AccordionItem
                id={`oidc-client4`}
                name="Client4"
                description="This is a temporary description!"
            ></AccordionItem>
            <AccordionItem
                id={`oidc-client4`}
                name="Client5"
                description="This is a temporary description!"
            ></AccordionItem>
        </Fragment>
    );
};

export default AdminView;
