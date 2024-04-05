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

import { ExistingScopes, OpenIDConnectClient } from "@models/OpenIDConnect";
import EditListItem from "@views/AdminUI/Common/EditListItem";
import MultiSelectDropdown from "@views/AdminUI/Common/MultiSelectDropdown";

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
    display: "flex",
    textAlign: "center",
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

const ClientAccordionSummary = styled(AccordionSummary)(({ theme }) => ({
    border: `1px solid ${theme.palette.divider}`,
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
            <ClientAccordionSummary
                expandIcon={<ArrowDropDownIcon />}
                aria-controls={`panel${props.client.ID}-content`}
                id={`panel${props.client.ID}-header`}
            >
                <div style={{ flex: 1, display: "flex", alignItems: "center", marginLeft: "16px" }}>
                    {isEditing ? (
                        <TextField
                            name="name"
                            value={formData.Name}
                            onChange={handleChange}
                            variant="outlined"
                            size="small"
                            onClick={(e) => e.stopPropagation()}
                        />
                    ) : (
                        <Typography fontWeight={"300"} fontSize={"20px"}>
                            {props.client.Name}
                        </Typography>
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
            </ClientAccordionSummary>
            <AccordionDetails sx={{ padding: "auto 16px" }} key={`accordion-details-${props.index}`}>
                <List>
                    <ListItem key={`client-id-${props.index}`}>
                        <Typography>{translate("Client ID: ") || "Client ID: "}</Typography>
                        {showClientID ? (
                            <CardArea elevation={0}>{props.client.ID}</CardArea>
                        ) : (
                            <CardArea elevation={0}>{"●".repeat(props.client.ID.length)}</CardArea>
                        )}
                        <IconButton onClick={toggleClientIDVisibility}>
                            {showClientID ? <VisibilityOffIcon /> : <VisibilityIcon />}
                        </IconButton>
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`client-type-${props.index}`}>Client Type: {props.client.ClientType}</ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`request-uris-${props.index}`}>
                        <List>
                            <Typography marginBottom={"0.5vh"}>{translate("Redirect URIs:  ")}</Typography>
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
                    <ListItem key={`scopes-${props.index}`}>
                        {isEditing ? (
                            <MultiSelectDropdown
                                index={props.index}
                                label={translate("Scopes:") || "Scopes:"}
                                values={formData.Scopes}
                                options={Object.values(ExistingScopes)}
                                handleChange={(updatedValues) => handleValuesUpdate(updatedValues, "Scopes")}
                            ></MultiSelectDropdown>
                        ) : (
                            <>
                                <Typography>
                                    {translate("Scopes: ")}
                                    {props.client.Scopes.join(", ")}
                                </Typography>
                            </>
                        )}
                    </ListItem>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`audience-${props.index}`}>
                        <List>
                            <Typography marginBottom={"0.5vh"}>{translate("Audience:") || "Audience:"}</Typography>
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
                    <ListItem key={`auth-policy-${props.index}`}>
                        {translate("Authorization Policy:")} {props.client.AuthorizationPolicy.Name}
                    </ListItem>
                </List>
            </AccordionDetails>
        </ClientAccordion>
    );
};

export default ClientItem;
