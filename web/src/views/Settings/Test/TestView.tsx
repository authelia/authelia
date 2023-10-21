import React, { Fragment, useState } from "react";

import { Button } from "@mui/material";
import TextField from "@mui/material/TextField";
import Grid from "@mui/material/Unstable_Grid2/Grid2";

import { deleteElevation, finishElevation, getElevation, startElevation } from "@services/Test";
import IdentityVerificationDialog from "@views/Settings/Common/IdentityVerificationDialog";

interface Props {}

const TestView = function (props: Props) {
    const [opening, setOpening] = useState<boolean>(false);
    const [otp, setOTP] = useState<string>("");
    const [deleteID, setDeleteID] = useState<string>("");

    const handleGet = async () => {
        const response = await getElevation();

        if (response === undefined) {
            return;
        }

        console.table(response);
    };

    const handleCreate = async () => {
        const response = await startElevation();

        console.table(response);
    };

    const handleSubmit = async () => {
        const response = await finishElevation(otp);

        console.log(response.data);
    };

    const handleDelete = async () => {
        const response = await deleteElevation(deleteID);

        console.log(response.data);
    };

    const handleClosed = (ok: boolean) => {
        console.log("handle close", ok);
        setOpening(false);
    };

    return (
        <Fragment>
            <IdentityVerificationDialog
                opening={opening}
                handleClosed={handleClosed}
                handleOpened={() => setOpening(false)}
            />
            <Grid container spacing={2}>
                <Grid xs={12}>
                    <Button onClick={() => setOpening(true)}>Show</Button>
                </Grid>
                <Grid xs={12}>
                    <Button onClick={handleGet}>Get</Button>
                </Grid>
                <Grid xs={12}>
                    <Button onClick={handleCreate}>Create</Button>
                </Grid>
                <Grid xs={12}>
                    <TextField
                        label={"One-Time Password"}
                        variant={"outlined"}
                        value={otp}
                        onChange={(e) => setOTP(e.target.value)}
                    />
                </Grid>
                <Grid xs={12}>
                    <Button onClick={handleSubmit}>Submit</Button>
                </Grid>
                <Grid xs={12}>
                    <TextField
                        label={"Delete ID"}
                        variant={"outlined"}
                        value={deleteID}
                        onChange={(e) => setDeleteID(e.target.value)}
                    />
                </Grid>
                <Grid xs={12}>
                    <Button onClick={handleDelete}>Delete</Button>
                </Grid>
            </Grid>
        </Fragment>
    );
};

export default TestView;
