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

const CardArea = styled(Paper)(({ theme }) => ({
    padding: "8px",
    borderRadius: "4px",
    fontFamily: "monospace",
    overflowX: "auto",
    [theme.breakpoints.down("sm")]: {
        width: "50vw",
    },
}));

const ClientItem = function (props: Props) {
    const { t: translate } = useTranslation("admin");
    const [isExpanded, setExpanded] = useState(false);
    const [showClientID, setShowClientID] = useState(false);
    //const theme = useTheme();

    const toggleExpanded = () => {
        setExpanded(!isExpanded);
    };

    const toggleClientIDVisibility = () => {
        setShowClientID((prevShowClientID) => !prevShowClientID);
    };

    const renderClientId = () => {
        if (showClientID) {
            return <CardArea elevation={0}>{props.client.ID}</CardArea>;
        } else {
            return <CardArea elevation={0}>{"●".repeat(props.client.ID.length)}</CardArea>;
        }
    };

    return (
        <Accordion
            expanded={isExpanded}
            onChange={toggleExpanded}
            sx={{
                width: "75vw",
                margin: "1vw auto", // default behavior sets left/right margin to '0' instead of auto, uncentering the accordion
                display: "flex",
                flexDirection: "column",
                "&.Mui-expanded": {
                    margin: "1vw auto",
                },
            }}
        >
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
                            {translate("Request URIs:")}
                            {props.client.RedirectURIs.map((uri, index) => (
                                <ListItem key={`request-uri-${props.index}-${index}`}>{uri}</ListItem>
                            ))}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`audience-${props.index}`}>
                        <List>
                            {translate("Audience:") || "Audience:"}
                            {props.client.Audience.map((audience, index) => (
                                <ListItem key={`audience-item-${index}`}>{audience}</ListItem>
                            ))}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`scopes-${props.index}`}>
                        <List>
                            {translate("Scopes:") || "Audience:"}
                            {props.client.Scopes.map((scopes, index) => (
                                <ListItem key={`scopes-item-${index}`}>{scopes}</ListItem>
                            ))}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`auth-policy-${props.index}`}>
                        {translate("Authorization Policy:")} {props.client.AuthorizationPolicy.Name}
                    </ListItem>
                </List>
            </AccordionDetails>
        </Accordion>
    );
};

export default ClientItem;
