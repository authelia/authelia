import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCheckCircle } from "@fortawesome/free-regular-svg-icons";

const SuccessIcon = function () {
    return (
        <FontAwesomeIcon icon={faCheckCircle} size="4x" color="green" className="success-icon" />
    )
}

export default SuccessIcon