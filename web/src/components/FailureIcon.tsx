import React from "react";

import { IconLookup } from "@fortawesome/fontawesome-svg-core";
import { faTimesCircle } from "@fortawesome/free-regular-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export interface Props {}

const FailureIcon = function (props: Props) {
    return <FontAwesomeIcon icon={faTimesCircle as IconLookup} size="4x" color="red" className="failure-icon" />;
};

export default FailureIcon;
