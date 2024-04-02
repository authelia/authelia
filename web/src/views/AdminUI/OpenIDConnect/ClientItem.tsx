import React, { useState } from "react";
//import React, { Fragment, useState } from "react";

//import { useTranslation } from "react-i18next";
import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";
import CloseIcon from "@mui/icons-material/Close";
import DoneAllIcon from "@mui/icons-material/DoneAll";
import EditIcon from "@mui/icons-material/Edit";
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
    TextField,
    Typography,
    styled,
} from "@mui/material";
import { useTranslation } from "react-i18next";

import { OpenIDConnectClient } from "@models/OpenIDConnect";
import EditListItem from "@views/AdminUI/Common/EditListItem";

interface Props {
    index: number;
    description: string;
    client: OpenIDConnectClient;
    handleChange: (index: number, updatedClient: OpenIDConnectClient) => void;
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

const ClientAccordion = styled(Accordion)(({ theme }) => ({
    width: "75vw",
    margin: "1vw auto", // default behavior sets left/right margin to '0' instead of auto, uncentering the accordion
    display: "flex",
    flexDirection: "column",
    "&.Mui-expanded": {
        margin: "1vw auto",
    },
}));

const ClientItem = function (props: Props) {
    const { t: translate } = useTranslation("admin");
    const [isExpanded, setExpanded] = useState(false);
    const [isEditing, setEditing] = useState(false);
    const [showClientID, setShowClientID] = useState(false);
    const [formData, setFormData] = useState<OpenIDConnectClient>(props.client);
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

    const handleEditClick = (event: { stopPropagation: () => void }) => {
        event.stopPropagation();
        if (!isExpanded) {
            toggleExpanded();
        }
        setEditing(true);
    };

    const handleSaveClick = (event: { stopPropagation: () => void }) => {
        event.stopPropagation();
        props.handleChange(props.index, formData);
        setEditing(false);
    };

    const handleStopEditClick = (event: { stopPropagation: () => void }) => {
        event.stopPropagation();
        setFormData(props.client);
        setEditing(false);
    };

    const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value } = e.target;
        console.log(`handleChange e: ${e}`);
        setFormData((prevData) => ({
            ...prevData,
            [name]: value,
        }));
    };

    const handleValuesUpdate = (updatedValues: string[], field: string) => {
        setFormData((prevData) => ({
            ...prevData,
            [field]: updatedValues,
        }));
    };

    return (
        <ClientAccordion expanded={isExpanded} onChange={toggleExpanded}>
            <AccordionSummary
                expandIcon={<ArrowDropDownIcon />}
                aria-controls={`panel${props.client.ID}-content`}
                id={`panel${props.client.ID}-header`}
            >
                <div style={{ flex: 1, display: "flex", alignItems: "center" }}>
                    {isEditing ? (
                        <TextField
                            name="name"
                            value={formData.Name}
                            onChange={handleChange}
                            variant="outlined"
                            size="small"
                        />
                    ) : (
                        <Typography>{props.client.Name}</Typography>
                    )}
                </div>
                <div style={{ display: "flex", alignItems: "center" }}>
                    {isEditing ? (
                        <>
                            <IconButton color={"success"} onClick={handleSaveClick}>
                                <DoneAllIcon />
                            </IconButton>
                            <IconButton color={"error"} onClick={handleStopEditClick}>
                                <CloseIcon />
                            </IconButton>
                        </>
                    ) : (
                        <IconButton onClick={handleEditClick}>
                            <EditIcon />
                        </IconButton>
                    )}
                </div>
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
                            {translate("Redirect URIs:  ")}
                            {isEditing ? (
                                <EditListItem
                                    values={formData.RedirectURIs}
                                    onValuesUpdate={(updatedValues) =>
                                        handleValuesUpdate(updatedValues, "RedirectURIs")
                                    }
                                />
                            ) : (
                                props.client.RedirectURIs.map((uri, index) => (
                                    <ListItem key={`request-uri-${props.index}-${index}`}>{uri}</ListItem>
                                ))
                            )}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`audience-${props.index}`}>
                        <List>
                            {translate("Audience:") || "Audience:"}
                            {isEditing ? (
                                <EditListItem
                                    values={formData.Audience}
                                    onValuesUpdate={(updatedValues) => handleValuesUpdate(updatedValues, "Audience")}
                                />
                            ) : (
                                props.client.Audience.map((audience, index) => (
                                    <ListItem key={`audience-item-${index}`}>{audience}</ListItem>
                                ))
                            )}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`scopes-${props.index}`}>
                        <List>
                            {translate("Scopes:") || "Scopes:"}
                            {isEditing ? (
                                <EditListItem
                                    values={formData.Scopes}
                                    onValuesUpdate={(updatedValues) => handleValuesUpdate(updatedValues, "Scopes")}
                                />
                            ) : (
                                props.client.Scopes.map((scopes, index) => (
                                    <ListItem key={`scopes-item-${index}`}>{scopes}</ListItem>
                                ))
                            )}
                        </List>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`auth-policy-${props.index}`}>
                        {translate("Authorization Policy:")} {props.client.AuthorizationPolicy.Name}
                    </ListItem>
                </List>
            </AccordionDetails>
        </ClientAccordion>
    );
};

export default ClientItem;
