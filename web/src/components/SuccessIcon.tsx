import { faCircleCheck } from "@fortawesome/free-regular-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/react-fontawesome";

const SuccessIcon = function () {
    return <FontAwesomeIcon icon={faCircleCheck} size="4x" color="green" className="success-icon" />;
};

export default SuccessIcon;
