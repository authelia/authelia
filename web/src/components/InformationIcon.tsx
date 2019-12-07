import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faInfoCircle } from "@fortawesome/free-solid-svg-icons";

export interface Props { }

export default function (props: Props) {
    return (
        <FontAwesomeIcon icon={faInfoCircle} size="4x" color="#5858ff" className="information-icon" />
    )
}