import React, { useState } from "react";
//import React, { Fragment, useState } from "react";

//import { useTranslation } from "react-i18next";
import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";
import VisibilityIcon from "@mui/icons-material/Visibility";
import VisibilityOffIcon from "@mui/icons-material/VisibilityOff";
import {
    Accordion,
    AccordionDetails,
    AccordionSummary,
    Divider,
    IconButton,
    List,
    ListItem,
    Paper,
    Typography,
    styled,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { OpenIDConnectClient } from "@models/OpenIDConnect";

interface Props {
    index: number;
    description: string;
    client: OpenIDConnectClient;
    handleInformation: (index: number) => void;
    handleEdit: (index: number) => void;
    handleDelete: (index: number) => void;
}

const CardArea = styled(Paper)({
    padding: "8px",
    backgroundColor: "#f0f0f0",
    borderRadius: "4px",
    fontFamily: "monospace",
});

const ClientItem = function (props: Props) {
    const { t: translate } = useTranslation("admin");

    //const [showDialogDetails, setShowDialogDetails] = useState<boolean>(false);
    const [isExpanded, setExpanded] = useState(false);
    const [showClientID, setShowClientID] = useState(false);

    const toggleExpanded = () => {
        setExpanded(!isExpanded);
    };

    const toggleClientIDVisibility = () => {
        setShowClientID((prevShowClientID) => !prevShowClientID);
    };

    // const renderClientId = () => {
    //     if (showClientID) {
    //         return  <MonospaceSpan>{props.client.ID}</MonospaceSpan>;
    //     } else {
    //         return <MonospaceSpan>{" "}{"●".repeat(props.client.ID.length)}</MonospaceSpan>;
    //     }
    // };
    const renderClientId = () => {
        if (showClientID) {
            return <CardArea>{props.client.ID}</CardArea>;
        } else {
            return <CardArea> {"●".repeat(props.client.ID.length)}</CardArea>;
        }
    };

    // default behavior sets left/right margin to '0' instead of auto, uncentering the accordion
    const CustomAccordion = styled(Accordion)(({ theme }) => ({
        width: "75%",
        margin: "16px auto",
        "&.Mui-expanded": {
            margin: "16px auto",
        },
    }));

    return (
        <CustomAccordion expanded={isExpanded} onChange={toggleExpanded}>
            <AccordionSummary
                expandIcon={<ArrowDropDownIcon />}
                aria-controls={`panel${props.client.ID}-content`}
                id={`panel${props.client.ID}-header`}
            >
                <Typography>{props.client.Name}</Typography>
            </AccordionSummary>
            <AccordionDetails key={`accordion-details-${props.index}`}>
                <List>
                    <ListItem key={`client-id-${props.index}`}>
                        {translate("Client ID:  ")}
                        {renderClientId()}
                        <IconButton onClick={toggleClientIDVisibility}>
                            {showClientID ? <VisibilityOffIcon /> : <VisibilityIcon />}
                        </IconButton>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`client-type-${props.index}`}>Client Type: {props.client.ClientType}</ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`request-uris-${props.index}`}>
                        <List>
                            Request URIs:{" "}
                            {props.client.RedirectURIs.map((uri, index) => (
                                <ListItem key={`request-uri-${props.index}-${index}`}>{uri}</ListItem>
                            ))}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`audience-${props.index}`}>
                        <List>
                            Audience:{" "}
                            {props.client.Audience.map((audience, index) => (
                                <ListItem key={`audience-item-${index}`}>{audience}</ListItem>
                            ))}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`scopes-${props.index}`}>
                        <List>
                            Scopes:{" "}
                            {props.client.Scopes.map((scopes, index) => (
                                <ListItem key={`scopes-item-${index}`}>{scopes}</ListItem>
                            ))}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`auth-policy-${props.index}`}>
                        Authorization Policy: {props.client.AuthorizationPolicy.Name}
                    </ListItem>
                </List>
            </AccordionDetails>
        </CustomAccordion>
    );
};

export default ClientItem;
