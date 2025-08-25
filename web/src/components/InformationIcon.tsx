import React from "react";

import { faCircleInfo } from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export interface Props {}

const InformationIcon = function (props: Props) {
    return <FontAwesomeIcon icon={faCircleInfo} size="4x" color="#5858ff" className="information-icon" />;
};

export default InformationIcon;
