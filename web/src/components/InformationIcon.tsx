// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

import React from "react";

import { faInfoCircle } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export interface Props {}

const InformationIcon = function (props: Props) {
    return <FontAwesomeIcon icon={faInfoCircle} size="4x" color="#5858ff" className="information-icon" />;
};

export default InformationIcon;
