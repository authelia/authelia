import React from "react";

import ArrowDropDownIcon from "@mui/icons-material/ArrowDropDown";
import { Accordion, AccordionDetails, AccordionSummary, Grid, Typography, styled } from "@mui/material";

interface Props {
    id: string;
    name: string;
    description: string;
}

// default behavior removed left/right margin when expanded (ie margin: 16px 0)
const CustomAccordion = styled(Accordion)(({ theme }) => ({
    width: "75%",
    margin: "16px auto",
    "&.Mui-expanded": {
        margin: "16px auto", // Override the expanded margin if needed
    },
}));

const AccordionItem = function (props: Props) {
    return (
        <CustomAccordion>
            <AccordionSummary
                expandIcon={<ArrowDropDownIcon />}
                aria-controls={`panel${props.id}-content`}
                id={`panel${props.id}-header`}
            >
                <Typography>{props.name}</Typography>
            </AccordionSummary>
            <AccordionDetails>
                <Grid container>
                    <Typography>{props.description}</Typography>
                </Grid>
            </AccordionDetails>
        </CustomAccordion>
    );
};

export default AccordionItem;
