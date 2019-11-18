import React from "react";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";
import { faCheckCircle } from "@fortawesome/free-regular-svg-icons";

export interface Props { }

export default function (props: Props) {
    return (
        <FontAwesomeIcon icon={faCheckCircle} size="4x" color="green" />
    )
}