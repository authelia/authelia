import { ReactNode, useState } from "react";

import PushNotificationIcon from "@components/PushNotificationIcon";
import { Button } from "@components/UI/Button";

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
                <div className="grid grid-cols-1 justify-center gap-2" id="device-selection">
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
                </div>
            );
            break;
        case State.METHOD:
            container = (
                <div className="grid grid-cols-1 justify-center gap-2" id="method-selection">
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
                </div>
            );
            break;
    }

    return (
        <div className="mx-auto flex max-w-lg flex-col items-center gap-2">
            {container}
            <Button variant={"ghost"} color={"primary"} onClick={props.onBack} id="device-selection-back">
                back
            </Button>
        </div>
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
        <div className={`${className} w-full`} id={idName}>
            <Button className="flex w-full flex-col items-center py-6" variant="default" onClick={props.onSelect}>
                <span className="fill-white">
                    <PushNotificationIcon width={32} height={32} />
                </span>
                <span className="text-xs">{props.device.name}</span>
            </Button>
        </div>
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
        <div className={`${className} w-full`} id={idName}>
            <Button className="flex w-full flex-col items-center py-6" variant="default" onClick={props.onSelect}>
                <span className="fill-white">
                    <PushNotificationIcon width={32} height={32} />
                </span>
                <span className="text-xs">{props.method}</span>
            </Button>
        </div>
    );
};

export default DefaultDeviceSelectionContainer;
