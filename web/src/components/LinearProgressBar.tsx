// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { cn } from "cn";

import { Progress } from "@components/UI/Progress";

export interface Props {
    value: number;
    height?: number | string;
    className?: string;
}

const LinearProgressBar = function (props: Props) {
    return (
        <Progress
            value={props.value}
            className={cn("mt-2 transition-transform duration-200 ease-linear", props.className)}
            style={{ height: props.height ?? 8 }}
        />
    );
};

export default LinearProgressBar;
