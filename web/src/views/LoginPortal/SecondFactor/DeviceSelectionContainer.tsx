import React, { ReactNode, useState } from "react";

import { Box, Button, Container, Grid, Theme, Typography } from "@mui/material";
import makeStyles from "@mui/styles/makeStyles";

import PushNotificationIcon from "@components/PushNotificationIcon";

export enum State {
    DEVICE = 1,
    METHOD = 2,
}

export interface SelectableDevice {
    id: string;
    name: string;
    methods: string[];
}

export interface SelectedDevice {
    id: string;
    method: string;
}

export interface Props {
    children?: ReactNode;
    devices: SelectableDevice[];

    onBack: () => void;
    onSelect: (device: SelectedDevice) => void;
}
const DefaultDeviceSelectionContainer = function (props: Props) {
    const [state, setState] = useState(State.DEVICE);
    const [device, setDevice] = useState([] as unknown as SelectableDevice);

    const handleDeviceSelected = (selecteddevice: SelectableDevice) => {
        if (selecteddevice.methods.length === 1) handleMethodSelected(selecteddevice.methods[0], selecteddevice.id);
        else {
            setDevice(selecteddevice);
            setState(State.METHOD);
        }
    };

    const handleMethodSelected = (method: string, deviceid?: string) => {
        if (deviceid) props.onSelect({ id: deviceid, method: method });
        else props.onSelect({ id: device.id, method: method });
    };

    let container: ReactNode;
    switch (state) {
        case State.DEVICE:
            container = (
                <Grid container justifyContent="center" spacing={1} id="device-selection">
                    {props.devices.map((value, index) => {
                        return (
                            <DeviceItem
                                id={index}
                                key={index}
                                device={value}
                                onSelect={() => handleDeviceSelected(value)}
                            />
                        );
                    })}
                </Grid>
            );
            break;
        case State.METHOD:
            container = (
                <Grid container justifyContent="center" spacing={1} id="method-selection">
                    {device.methods.map((value, index) => {
                        return (
                            <MethodItem
                                id={index}
                                key={index}
                                method={value}
                                onSelect={() => handleMethodSelected(value)}
                            />
                        );
                    })}
                </Grid>
            );
            break;
    }

    return (
        <Container>
            {container}
            <Button color="primary" onClick={props.onBack} id="device-selection-back">
                back
            </Button>
        </Container>
    );
};

interface DeviceItemProps {
    id: number;
    device: SelectableDevice;

    onSelect: () => void;
}

const DeviceItem = function (props: DeviceItemProps) {
    const className = "device-option-" + props.id;
    const idName = "device-" + props.device.id;
    const style = makeStyles((theme: Theme) => ({
        item: {
            paddingTop: theme.spacing(4),
            paddingBottom: theme.spacing(4),
            width: "100%",
        },
        icon: {
            display: "inline-block",
            fill: "white",
        },
        buttonRoot: {
            display: "block",
        },
    }))();

    return (
        <Grid item xs={12} className={className} id={idName}>
            <Button
                className={style.item}
                color="primary"
                classes={{ root: style.buttonRoot }}
                variant="contained"
                onClick={props.onSelect}
            >
                <Box className={style.icon}>
                    <PushNotificationIcon width={32} height={32} />
                </Box>
                <Box>
                    <Typography>{props.device.name}</Typography>
                </Box>
            </Button>
        </Grid>
    );
};

interface MethodItemProps {
    id: number;
    method: string;

    onSelect: () => void;
}

const MethodItem = function (props: MethodItemProps) {
    const className = "method-option-" + props.id;
    const idName = "method-" + props.method;
    const style = makeStyles((theme: Theme) => ({
        item: {
            paddingTop: theme.spacing(4),
            paddingBottom: theme.spacing(4),
            width: "100%",
        },
        icon: {
            display: "inline-block",
            fill: "white",
        },
        buttonRoot: {
            display: "block",
        },
    }))();

    return (
        <Grid item xs={12} className={className} id={idName}>
            <Button
                className={style.item}
                color="primary"
                classes={{ root: style.buttonRoot }}
                variant="contained"
                onClick={props.onSelect}
            >
                <Box className={style.icon}>
                    <PushNotificationIcon width={32} height={32} />
                </Box>
                <Box>
                    <Typography>{props.method}</Typography>
                </Box>
            </Button>
        </Grid>
    );
};

export default DefaultDeviceSelectionContainer;
