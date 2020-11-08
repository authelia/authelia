import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faTimesCircle } from "@fortawesome/free-regular-svg-icons";

export interface Props { }

const FailureIcon = function (props: Props) {
    return (
        <FontAwesomeIcon icon={faTimesCircle} size="4x" color="red" className="failure-icon" />
    )
}

export default FailureIcon