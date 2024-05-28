import React, { Fragment, useState } from "react";

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
    client: OpenIDConnectClient;
    handleChange: (index: number, updatedClient: OpenIDConnectClient) => void;
    handleDelete: (index: number) => void;
}

const CardArea = styled(Paper)(({ theme }) => ({
    padding: "6px",
    fontFamily: "monospace",
    display: "inline-block",
    textAlign: "center",
    [theme.breakpoints.down("sm")]: {
        width: "50vw",
    },
}));

const ClientAccordion = styled(Accordion)(({ theme }) => ({
    width: "75vw",
    margin: "8px auto", // default behavior sets left/right margin to '0' instead of auto, uncentering the accordion
    display: "flex",
    flexDirection: "column",
    "&.Mui-expanded": {
        margin: "8px auto",
    },
    [theme.breakpoints.down("sm")]: {
        width: "100%",
        margin: 0,
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

    const toggleExpanded = () => {
        setExpanded(!isExpanded);
    };

    const toggleClientIDVisibility = () => {
        setShowClientID((prevShowClientID) => !prevShowClientID);
    };

    const handleEditClick = (event: { stopPropagation: () => void }) => {
        event.stopPropagation();
        setFormData(props.client);
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
                            sx={{ fontWeight: "300", fontSize: "20px" }}
                            onClick={(e: { stopPropagation: () => any }) => e.stopPropagation()}
                        />
                    ) : (
                        <Typography fontWeight={"300"} fontSize={"20px"} component="div">
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
                        <div>
                            {translate("Client ID: ") || "Client ID: "}
                            {showClientID ? (
                                <CardArea>{props.client.ID}</CardArea>
                            ) : (
                                <CardArea>{"●".repeat(props.client.ID.length)}</CardArea>
                            )}
                            <IconButton onClick={toggleClientIDVisibility}>
                                {showClientID ? <VisibilityOffIcon /> : <VisibilityIcon />}
                            </IconButton>
                        </div>
                    </ListItem>
                    <Fragment>
                        <Divider variant="middle" component="li" />
                        <ListItem key={`client-type-${props.index}`}>
                            Client Type: {props.client.Public ? `Public` : `Confidential`}
                        </ListItem>
                    </Fragment>
                    <Divider variant="middle" component="li" />
                    <ListItem key={`request-uris-${props.index}`}>
                        <List sx={{ width: "50%", padding: 0 }}>
                            <Typography marginBottom={"0.5vh"}>{translate("Redirect URIs:  ")}</Typography>
                            {isEditing ? (
                                <EditListItem
                                    listLabel={`RedirectURIs`}
                                    index={props.index}
                                    values={formData.RedirectURIs}
                                    onValuesUpdate={(updatedValues) =>
                                        handleValuesUpdate(updatedValues, "RedirectURIs")
                                    }
                                />
                            ) : (
                                props.client.RedirectURIs.map((uri, index) => (
                                    <ListItem
                                        sx={{ paddingTop: 0, paddingBottom: 0 }}
                                        key={`redirect-uri-${props.index}-${index}`}
                                    >
                                        {uri}
                                    </ListItem>
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
                    {props.client.Audience && (
                        <Fragment>
                            <Divider variant="middle" component="li" />
                            <ListItem key={`audience-${props.index}`}>
                                <List>
                                    <Typography marginBottom={"0.5vh"}>
                                        {translate("Audience:") || "Audience:"}
                                    </Typography>
                                    {isEditing ? (
                                        <EditListItem
                                            index={props.index}
                                            listLabel={`Audience`}
                                            values={formData.Audience ?? []}
                                            onValuesUpdate={(updatedValues) =>
                                                handleValuesUpdate(updatedValues, "Audience")
                                            }
                                        />
                                    ) : (
                                        props.client.Audience.map((audience, index) => (
                                            <ListItem key={`audience-item-${index}`}>{audience}</ListItem>
                                        ))
                                    )}
                                </List>
                            </ListItem>
                        </Fragment>
                    )}
                    {props.client.AuthorizationPolicy && (
                        <Fragment>
                            <Divider variant="middle" component="li" />
                            <ListItem key={`auth-policy-${props.index}`}>
                                {translate("Authorization Policy:")} {props.client.AuthorizationPolicy.Name}
                            </ListItem>
                        </Fragment>
                    )}
                </List>
            </AccordionDetails>
        </ClientAccordion>
    );
};

export default ClientItem;
