// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import { AlertColor } from "@mui/material";

export interface Notification {
    message: string;
    level: AlertColor;
    timeout: number;
}
