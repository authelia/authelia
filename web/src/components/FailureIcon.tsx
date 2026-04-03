import { faCircleXmark } from "@fortawesome/free-regular-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

export interface Props {}

const FailureIcon = function () {
    return <FontAwesomeIcon icon={faCircleXmark} size="4x" color="red" className="failure-icon" />;
};

export default FailureIcon;
