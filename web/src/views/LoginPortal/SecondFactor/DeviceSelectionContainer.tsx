import { ReactNode, useState } from "react";

import { Box, Button, Container, Typography } from "@mui/material";
import Grid from "@mui/material/Grid";

import PushNotificationIcon from "@components/PushNotificationIcon";

/* eslint-disable no-unused-vars */
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
    devices: SelectableDevice[];

    onBack: () => void;
    onSelect: (_device: SelectedDevice) => void;
}
const DefaultDeviceSelectionContainer = function (props: Props) {
    const [state, setState] = useState(State.DEVICE);
    const [device, setDevice] = useState([] as unknown as SelectableDevice);

    const handleDeviceSelected = (selectedDevice: SelectableDevice) => {
        if (selectedDevice.methods.length === 1) handleMethodSelected(selectedDevice.methods[0], selectedDevice.id);
        else {
            setDevice(selectedDevice);
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
                                key={value.id}
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
                                key={value}
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

    return (
        <Grid size={{ xs: 12 }} className={className} id={idName}>
            <Button
                sx={{
                    display: "block",
                    paddingBottom: (theme) => theme.spacing(4),
                    paddingTop: (theme) => theme.spacing(4),
                    width: "100%",
                }}
                color="primary"
                variant="contained"
                onClick={props.onSelect}
            >
                <Box sx={{ display: "inline-block", fill: "white" }}>
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

    return (
        <Grid size={{ xs: 12 }} className={className} id={idName}>
            <Button
                sx={{
                    display: "block",
                    paddingBottom: (theme) => theme.spacing(4),
                    paddingTop: (theme) => theme.spacing(4),
                    width: "100%",
                }}
                color="primary"
                variant="contained"
                onClick={props.onSelect}
            >
                <Box sx={{ display: "inline-block", fill: "white" }}>
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
