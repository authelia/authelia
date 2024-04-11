import React, { useState } from "react";

import CheckIcon from "@mui/icons-material/Check";
import CloseIcon from "@mui/icons-material/Close";
import { IconButton, List, ListItem, TextField } from "@mui/material";

interface Props {
    index: number;
    listLabel: string;
    values: string[];
    onValuesUpdate: (updatedValues: string[]) => void;
}

const EditListItem = (props: Props) => {
    //const [editedValues, setEditedValues] = useState<string[]>(props.values);
    const [newFormValues, setNewFormValues] = useState<string[]>(props.values);
    const [newFieldValue, setNewFieldValue] = useState<string>("");
    console.log(newFormValues);

    const handleInputChange = (index: number, newValue: string) => {
        const updatedValues = [...newFormValues];
        updatedValues[index] = newValue;
        setNewFormValues(updatedValues);
        props.onValuesUpdate(updatedValues);
    };

    const handleDelete = (index: number) => {
        const updatedValues = [...newFormValues];
        const filteredValues = updatedValues.filter((_: any, i: any) => i !== index);
        setNewFormValues(filteredValues);
        props.onValuesUpdate(filteredValues);
    };

    const handleAddField = () => {
        if (newFieldValue.trim() !== "") {
            const updatedValues = [...newFormValues, newFieldValue];
            setNewFormValues(updatedValues);
            props.onValuesUpdate(updatedValues);
            setNewFieldValue("");
        }
    };

    return (
        <div>
            <List>
                {newFormValues.map((value, index) => (
                    <ListItem key={`edit-${props.listLabel}-${props.index}-${index}`}>
                        <TextField
                            fullWidth
                            size="small"
                            value={value}
                            onChange={(event: { target: { value: string } }) =>
                                handleInputChange(index, event.target.value)
                            }
                        />
                        <IconButton onClick={() => handleDelete(index)}>
                            <CloseIcon color={"error"} />
                        </IconButton>
                    </ListItem>
                ))}
                <ListItem key={`add-value-${props.listLabel}-${props.index}`}>
                    <TextField
                        size="small"
                        onChange={(event: { target: { value: string } }) => setNewFieldValue(event.target.value)}
                        value={newFieldValue}
                        placeholder="New Value"
                    />
                    <IconButton onClick={handleAddField}>
                        <CheckIcon color={"success"} />
                    </IconButton>
                </ListItem>
            </List>
        </div>
    );
};

export default EditListItem;
