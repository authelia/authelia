import React, { useEffect, useState } from "react";

import CheckBoxIcon from "@mui/icons-material/CheckBox";
import CheckBoxOutlineBlankIcon from "@mui/icons-material/CheckBoxOutlineBlank";
import { Autocomplete, Checkbox, TextField } from "@mui/material";

interface Props {
    index: number;
    label: string;
    values: string[];
    options: string[];
    handleChange: (updatedValues: string[]) => void;
}

const icon = <CheckBoxOutlineBlankIcon fontSize="small" />;
const checkedIcon = <CheckBoxIcon fontSize="small" />;

const MultiSelectDropdown = function (props: Props) {
    const [formValues, setFormValues] = useState<string[]>(props.values);

    const handleChange = (newValues: string[]) => {
        setFormValues(newValues);
        props.handleChange(newValues);
    };

    useEffect(() => {
        setFormValues(props.values);
    }, [props.values]);

    return (
        <div>
            <Autocomplete
                multiple
                id="checkboxes-tags-demo"
                options={props.options}
                disableCloseOnSelect
                getOptionLabel={(option) => option}
                renderOption={(props, option, { selected }) => (
                    <li {...props}>
                        <Checkbox icon={icon} checkedIcon={checkedIcon} style={{ marginRight: 8 }} checked={selected} />
                        {option}
                    </li>
                )}
                style={{ width: 500 }}
                renderInput={(params) => (
                    <TextField
                        {...params}
                        value={formValues}
                        label={props.label}
                        onChange={(event) => handleChange(Array.isArray(event.target.value) ? event.target.value : [])}
                    />
                )}
                value={formValues}
                onChange={(event, newValue) => handleChange(Array.isArray(newValue) ? newValue : [])}
            />
        </div>
    );
};

export default MultiSelectDropdown;
