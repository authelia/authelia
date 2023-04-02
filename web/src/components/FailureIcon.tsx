// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import React from "react";

import { faTimesCircle } from "@fortawesome/free-regular-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export interface Props {}

const FailureIcon = function (props: Props) {
    return <FontAwesomeIcon icon={faTimesCircle} size="4x" color="red" className="failure-icon" />;
};

export default FailureIcon;
